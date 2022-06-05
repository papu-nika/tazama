package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-runewidth"
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

type SearchStr []struct {
	str string
}

type Options struct {
	op_v bool
}

var file_buf FileBuf = map[int]string{}
var scroll Scroll = Scroll{0, 0}
var search_strs SearchStr = SearchStr{{}}

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

	cache_PollEvent, callDrawTerm_ch := create_cache_poolEvent()

	index_buf.Chach_input(cache_PollEvent, callDrawTerm_ch)
}

func create_cache_poolEvent() (<-chan termbox.Event, <-chan interface{}) {
	pollEvent_ch := make(chan termbox.Event, 100)
	callDrawTerm_ch := make(chan interface{})
	go func() {
		for {
			ev := termbox.PollEvent()
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				termbox.Close()
				os.Exit(0)
			case termbox.KeyArrowRight, termbox.KeyArrowLeft, termbox.KeyArrowUp, termbox.KeyArrowDown:
				key_arrow_procces(ev.Key)
				callDrawTerm_ch <- make(chan interface{})
			case termbox.KeySpace, termbox.KeyBackspace2:
				pollEvent_ch <- ev
			default:
				if ev.Ch == 92 {
					search_strs[0].str += "\\"
					continue
				} else {
					search_strs[0].str += string(ev.Ch)
				}
				promptPrint()
				termbox.Flush()
				log.Println("default acction")
				pollEvent_ch <- ev
			}
		}
	}()
	return pollEvent_ch, callDrawTerm_ch
}

func promptPrint() {
	log.Println("Prompt prrint2")
	var prompt rune = '>'
	var str string

	for _, v := range search_strs {
		str = v.str + " " + str
	}
	str = str[:len(str)-1]
	str_rune := []rune(str)

	termbox.SetCell(0, 0, prompt, default_color, default_color)
	x := 2
	for i := 0; i < len(str_rune); i++ {
		termbox.SetCell(x, 0, rune(str_rune[i]), default_color, default_color)
		x += runewidth.RuneWidth(rune(str_rune[i]))
	}
	termbox.SetCell(x, 0, ' ', default_color, termbox.ColorWhite)
}

func (index_buf *IndexBuf) Chach_input(cache_PollEvent <-chan termbox.Event, callDrawTerm_ch <-chan interface{}) {
	cache_DrawTerm := func(
		done <-chan interface{},
		callDrawTerm_ch <-chan interface{},
	) <-chan interface{} {
		done_DrawTerm_ch := make(chan interface{})
		go func(index_buf *IndexBuf) {
			defer close(done_DrawTerm_ch)
			defer log.Println("draw done!!")
			for {
				select {
				case <-callDrawTerm_ch:
					log.Println("callDrawTerm")
					index_buf.Draw_Termbox()
					promptPrint()
					termbox.Flush()
				case <-done:
					return
				}
			}
		}(index_buf)
		return done_DrawTerm_ch
	}

	done := make(chan interface{})
	done_DrawTerm_ch := cache_DrawTerm(done, callDrawTerm_ch)

MAINLOOP:
	for {
		index_buf.Draw_Termbox()
		search_strs.PromptPrint()
		termbox.Flush()
	BUFFER_RELATED_WITHOUT:
		for {
			termbox.Flush()
			ev := <-cache_PollEvent
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				// case termbox.KeyArrowRight, termbox.KeyArrowLeft, termbox.KeyArrowUp, termbox.KeyArrowDown:
				// 	key_arrow_procces(ev.Key)
				// 	index_buf.Draw_Termbox()
				// 	search_strs.PromptPrint()
				// 	continue BUFFER_RELATED_WITHOUT
				case termbox.KeyBackspace2:
					if len(search_strs) == 1 && search_strs[0].str == "" {
						continue BUFFER_RELATED_WITHOUT
					} else if search_strs[0].str == "" {
						search_strs = (search_strs)[1:]
						close(done)
						<-done_DrawTerm_ch
						return
					} else {
						(search_strs)[0].str = search_strs[0].str[:len(search_strs[0].str)-1]
						continue MAINLOOP
					}
				case termbox.KeySpace:
					if search_strs[0].str == "" {
						continue BUFFER_RELATED_WITHOUT
					} else if search_strs[0].str[len(search_strs[0].str)-1] == '\\' {
						search_strs[0].str += " "
						continue BUFFER_RELATED_WITHOUT
					}
					if error_message := is_ok_regex(search_strs[0].str); error_message != "" {
						error_print(error_message)
						continue BUFFER_RELATED_WITHOUT
					}
					search_strs = append(SearchStr{{""}}, search_strs...)
					promptPrint()
					termbox.Flush()
					new_index_buf := index_buf.Re_Create_buf()
					log.Println(len(cache_PollEvent))
					close(done)
					<-done_DrawTerm_ch
					log.Println("kokoniitteiruka")
					new_index_buf.Chach_input(cache_PollEvent, callDrawTerm_ch)
					log.Println("remake done")
					done = make(chan interface{})
					done_DrawTerm_ch = cache_DrawTerm(done, callDrawTerm_ch)
					continue MAINLOOP
				default:
					// if ev.Ch == 92 {
					// 	search_strs[0].str += "\\"
					// 	continue BUFFER_RELATED_WITHOUT
					// } else {
					// 	search_strs[0].str += string(ev.Ch)
					// }

					// search_strs.PromptPrint()
					continue BUFFER_RELATED_WITHOUT
				}
			}
		}
	}
}

func (index_buf *IndexBuf) Re_Create_buf() *IndexBuf {
	re_now := time.Now()
	var new_index_buf IndexBuf = IndexBuf{}

	if search_strs[1].str == "sort" {
		index_buf.New_Srot_buf(&new_index_buf)

	} else if search_strs[1].str == "uniq" || search_strs[1].str == "uniq-c" {
		index_buf.New_Uniq_buf(&new_index_buf)
	} else if search_strs[1].str == "cut" {
		index_buf.New_Cut_buf(&new_index_buf, 3)
	} else {
		index_buf.New_Grep_buf(&new_index_buf)
	}
	log.Printf("##Re_Create_buf##\t%d milisecond\tkey= \"%s\"", time.Since(re_now).Milliseconds(), search_strs[1].str)
	//log.Printf("%p, new=%p", index_buf, &new_index_buf)
	// for key, value := range new_index_buf {
	// 	log.Printf("key %d = value %d", key, value.cut_range)
	// }
	return &new_index_buf
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
			search_strs.PromptPrint()
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
