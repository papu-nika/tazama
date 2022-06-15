package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	"golang.org/x/crypto/ssh/terminal"
)

const version = "0.0.1"

const (
	default_color = termbox.ColorDefault
	warning_color = termbox.ColorRed
	match_color   = termbox.ColorBlue
)

// int=行数,string=文字列
type FileBuf map[int]string

// int=マッチした行,Keys=属性
type IndexBuf map[int]Keys

type Keys struct {
	index      int
	uniq_num   int
	grep_range []int
	cut_range  []int
}

type Scroll struct {
	high  int
	width int
}

type PromptStrs []struct {
	str string
}

type Options struct {
	op_v bool
}

type Wordprocces struct {
	word string
}

var file_buf FileBuf = map[int]string{}
var scroll Scroll = Scroll{0, 0}

type CallDrawCh struct {
	draw <-chan interface{}
}

type CallEventCh struct {
	event <-chan termbox.Event
}

type CallPoolCh struct {
	word   <-chan string
	delete <-chan interface{}
}

type CallTazamaCh struct {
	do      <-chan string
	done    chan<- interface{}
	delete  <-chan interface{}
	deleted chan<- interface{}
}

func main() {
	log_file := log_init()
	defer log_file.Close()
	file := args_process()
	if err := termbox.Init(); err != nil {
		is_error(err)
	}
	defer termbox.Close()
	index_buf := read_file(file)
	file.Close()

	callEventCh, _ := cacheEvent()
	callPoolCh := poolEvent(callEventCh)
	callTazamaCh := controlEvent(callPoolCh)

	var index_bufs []*IndexBuf
	index_bufs = append(index_bufs, index_buf)
	doTazama(index_bufs, callTazamaCh)

}

// Ecs, Ctrl+C, KeyArrowはここで処理して、他のキーは流す。
func cacheEvent() (CallEventCh, CallDrawCh) {
	callEventCh := make(chan termbox.Event, 100)
	callDrawCh := make(chan interface{})
	var callEventCh_re CallEventCh = CallEventCh{event: callEventCh}
	var callDrawCh_re CallDrawCh = CallDrawCh{draw: callDrawCh}

	go func() {
		for {
			ev := termbox.PollEvent()
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				termbox.Close()
				os.Exit(0)
			case termbox.KeyArrowRight, termbox.KeyArrowLeft, termbox.KeyArrowUp, termbox.KeyArrowDown:
				key_arrow_procces(ev.Key)
				callDrawCh <- make(chan interface{})
			default:
				callEventCh <- ev
			}
		}
	}()
	return callEventCh_re, callDrawCh_re
}

func poolEvent(callEventCh CallEventCh) CallPoolCh {
	callWordCh := make(chan string)
	callDeleteCh := make(chan interface{})
	var callPoolCh_re CallPoolCh = CallPoolCh{
		word:   callWordCh,
		delete: callDeleteCh,
	}

	var promptStrs PromptStrs = PromptStrs{{}}
	go func(callEventCh CallEventCh, callWordCh chan string, callDeleteCh chan interface{}) {
		for {
			promptStrs.PromptPrint()
			ev := <-callEventCh.event
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyBackspace2:
					if len(promptStrs) == 1 && promptStrs[0].str == "" {
						continue
					} else if promptStrs[0].str == "" {
						callDeleteCh <- (promptStrs)[1].str
						promptStrs = (promptStrs)[1:]
						continue
					} else {
						(promptStrs)[0].str = promptStrs[0].str[:len(promptStrs[0].str)-1]
						continue
					}
				case termbox.KeySpace:
					if promptStrs[0].str == "" {
						continue
					} else if promptStrs[0].str[len(promptStrs[0].str)-1] == '\\' {
						promptStrs[0].str += " "
						continue
					} else if error_message := is_ok_regex(promptStrs[0].str); error_message != "" {
						error_print(error_message)
						continue
					} else {
						promptStrs = append(PromptStrs{{""}}, promptStrs...)
						callWordCh <- promptStrs[1].str
						continue
					}
				default:
					if ev.Ch == 92 {
						promptStrs[0].str += "\\"
						continue
					} else {
						promptStrs[0].str += string(ev.Ch)
					}
				}
			}
		}
	}(callEventCh, callWordCh, callDeleteCh)
	return callPoolCh_re
}

func controlEvent(callWordCh CallPoolCh) CallTazamaCh {
	doTazamaCh := make(chan string)
	doneTazamaCh := make(chan interface{})
	deleteTamazaCh := make(chan interface{})
	deletedTamazaCh := make(chan interface{})
	var callTazamaCh CallTazamaCh = CallTazamaCh{
		do:      doTazamaCh,
		done:    doneTazamaCh,
		delete:  deleteTamazaCh,
		deleted: deletedTamazaCh,
	}

	doPoolCh := callWordCh.word
	deletePoolCh := callWordCh.delete

	var wordbuf []string = []string{}
	var lock sync.Mutex

	// Word Relay
	go func(wordbuf *[]string) {
		wordRelay := func() {
			w := <-doPoolCh
			lock.Lock()
			defer lock.Unlock()
			*wordbuf = append(*wordbuf, w)
		}
		for {
			wordRelay()
		}
	}(&wordbuf)

	// Detele Relay
	go func(wordbuf *[]string) {
		deleteRelay := func() {
			w := <-deletePoolCh
			log.Println(w)
			lock.Lock()
			defer lock.Unlock()
			if len(*wordbuf) > 0 {
				*wordbuf = (*wordbuf)[:len(*wordbuf)-1]
			} else if len(*wordbuf) == 0 {
				deleteTamazaCh <- make(chan interface{})
			}
			<-deletedTamazaCh
		}
		for {
			deleteRelay()
		}
	}(&wordbuf)

	sendTazamaWord := func(wordbuf *[]string) {
		lock.Lock()
		defer lock.Unlock()
		doTazamaCh <- (*wordbuf)[len(*wordbuf)-1]
		*wordbuf = (*wordbuf)[:len(*wordbuf)-1]
		<-doneTazamaCh
	}

	go func(wordbuf *[]string) {
		for {
			if len(*wordbuf) > 0 {
				sendTazamaWord(wordbuf)
			}
		}
	}(&wordbuf)
	return callTazamaCh
}

func doTazama(index_bufs []*IndexBuf, callTazamaCh CallTazamaCh) {
	doCh := callTazamaCh.do
	doneCh := callTazamaCh.done
	deleteCh := callTazamaCh.delete
	deletedCh := callTazamaCh.deleted

	getNewIndex := func(word string, new_index_buf *IndexBuf) {
		index_bufs[len(index_bufs)-1].Re_Create_buf(word, new_index_buf)
	}

	for {
		if len(index_bufs) == 0 {
			log.Println("error: index_buf is no indexs")
			os.Exit(1)
		}
		index_bufs[len(index_bufs)-1].Draw_Termbox()
		termbox.Flush()

		var wg sync.WaitGroup

		select {
		case word := <-doCh:
			var new_index_buf IndexBuf = IndexBuf{}
			wg.Add(1)
			go func(deleteCh <-chan interface{}, new_index_buf *IndexBuf, doneCh chan<- interface{}, word string) {
				defer func() {
					wg.Done()
					doneCh <- make(chan interface{})
				}()
				getNewIndex(word, new_index_buf)
			}(deleteCh, &new_index_buf, doneCh, word)
			wg.Wait()
			index_bufs = append(index_bufs, &new_index_buf)
		case <-deleteCh:
			index_bufs = index_bufs[:len(index_bufs)-1]
			deletedCh <- make(chan interface{})
		}
	}
}

func (index_buf *IndexBuf) Re_Create_buf(str string, new_index_buf *IndexBuf) {
	re_now := time.Now()

	if str == "sort" {
		index_buf.New_Srot_buf(new_index_buf)
	} else if str == "uniq" || str == "uniq-c" {
		index_buf.New_Uniq_buf(new_index_buf)
	} else if str == "cut" {
		index_buf.New_Cut_buf(new_index_buf, 3)
	} else {
		index_buf.New_Grep_buf(new_index_buf, str)
	}
	log.Printf("##Re_Create_buf##\t%d milisecond\tkey= \"%s\"", time.Since(re_now).Milliseconds(), str)
}

func is_ok_regex(str string) string {
	_, err := regexp.CompilePOSIX(str)
	if err != nil {
		return "不正な正規表現です"
	} else {
		return ""
	}
}

func key_arrow_procces(ev termbox.Key) {
	switch ev {
	case termbox.KeyArrowDown:
		scroll.high++
	case termbox.KeyArrowUp:
		if scroll.high == 0 {
			return
		} else {
			scroll.high--
		}
	case termbox.KeyArrowRight:
		scroll.width++
	case termbox.KeyArrowLeft:
		if scroll.width == 0 {
			return
		}
		scroll.width--
	}
	return
}

func args_process() *os.File {
	if terminal.IsTerminal(0) && len(os.Args) == 1 {
		fmt.Println("ファイルを引数に指定するか、パイプで標準入力を与えてください")
		os.Exit(1)
	}
	option := Options{}
	flag.BoolVar(&option.op_v, "v", false, "show version")
	flag.BoolVar(&option.op_v, "version", false, "show version")
	flag.Parse()
	if option.op_v {
		fmt.Println("version: ", version)
		os.Exit(0)
	}
	argments := flag.Args()
	if len(os.Args) > 1 {
		f, err := os.Open(argments[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return f
	} else {
		f, err := os.Open(os.Stdin.Name())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return f
	}
}

func read_file(f *os.File) *IndexBuf {
	now := time.Now()
	var index_buf IndexBuf = map[int]Keys{}

	_, high_size := termbox.Size()

	read := bufio.NewScanner(f)
	var i int
	for i = 1; read.Scan(); i++ {
		file_buf[i] = read.Text()

		index_buf[i] = Keys{
			index:      i,
			uniq_num:   0,
			grep_range: nil,
			cut_range:  []int{0, len([]rune(file_buf[i]))},
		}
		if i == high_size {
			index_buf[0] = Keys{
				index:      i,
				uniq_num:   0,
				grep_range: nil,
				cut_range:  nil,
			}
			index_buf.Draw_Termbox()
			termbox.Flush()
		}
	}
	index_buf[0] = Keys{i, 0, nil, nil}
	log.Printf("##Read_File##\t%d milisecond\tkey= \"%s\"", time.Since(now).Milliseconds(), "")
	return &index_buf
}

func (index_buf *IndexBuf) Check_key(index_key int) bool {
	_, ok := (*index_buf)[index_key]
	if ok == false {
		return false
	}
	return true
}

func log_init() *os.File {
	file, err := os.OpenFile("_test.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		is_error(err)
	}
	log.SetOutput(file)
	return file
}

func is_error(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func error_proccess(err error) {
	termbox.Close()
	log.Print(err)
	os.Exit(1)
}
