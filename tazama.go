package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

const coldef = termbox.ColorDefault

type File_buf map[uint32]string

type Key struct {
	index uint32
	match []int
}

type Index_buf map[uint32]Key

type Search_strs struct {
	str      string
	bg_color termbox.Attribute
	fg_color termbox.Attribute
}

func main() {
	var buf File_buf = map[uint32]string{}
	var index_buf Index_buf = map[uint32]Key{}
	var high_scroll uint32 = 0
	var width_scroll int = 0
	var search_strs []Search_strs = []Search_strs{{"", 0, 0}}

	file := log_init()
	defer file.Close()
	flag.Parse()
	argments := flag.Args()

	if err := termbox.Init(); err != nil {
		is_error(err)
	}
	defer termbox.Close()
	(&buf).Read_File(argments[0], &index_buf)
	(&buf).Chach_input(index_buf, &high_scroll, &width_scroll, &search_strs)
}

func (buf *File_buf) Chach_input(index_buf Index_buf, high_scroll *uint32, width_scroll *int, search_strs *[]Search_strs) {
MAINLOOP:
	for {
		buf.Drow_Termbox(*high_scroll, *width_scroll, search_strs, &index_buf)
	BUFFER_RELATED_WITHOUT:
		for {
			ev := termbox.PollEvent()
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					termbox.Close()
					os.Exit(0)
				case termbox.KeyArrowRight, termbox.KeyArrowLeft, termbox.KeyArrowUp, termbox.KeyArrowDown:
					buf.Keyarrow_procces(high_scroll, width_scroll, ev.Key, &index_buf, search_strs)
					continue BUFFER_RELATED_WITHOUT
				case termbox.KeyBackspace2:
					if len(*search_strs) == 1 && (*search_strs)[0].str == "" {
						continue BUFFER_RELATED_WITHOUT
					} else if (*search_strs)[0].str == "" {
						*search_strs = (*search_strs)[1:]
						return
					} else {
						(*search_strs)[0].str = (*search_strs)[0].str[:len((*search_strs)[0].str)-1]
						continue MAINLOOP
					}
				case termbox.KeySpace:
					if (*search_strs)[0].str == "" {
						continue MAINLOOP
					} else if (*search_strs)[0].str[len((*search_strs)[0].str)-1] == '\\' {
						(*search_strs)[0].str += " "
						continue MAINLOOP
					}
					*search_strs = append([]Search_strs{{"", 0, 0}}, (*search_strs)...)

					new_index_buf := *buf.Re_Create_buf(*high_scroll, *width_scroll, search_strs, &index_buf)

					buf.Chach_input(new_index_buf, high_scroll, width_scroll, search_strs)
					continue MAINLOOP
				default:
					if ev.Ch == 92 {
						(*search_strs)[0].str += "\\"
						buf.Drow_Termbox(*high_scroll, *width_scroll, search_strs, &index_buf)
						continue BUFFER_RELATED_WITHOUT
					} else {
						(*search_strs)[0].str += string(ev.Ch)
					}
					//buf.Chach_input(*buf.Re_Create_buf(*high_scroll, *width_scroll, search_strs, &index_buf), high_scroll, width_scroll, search_strs)
					continue MAINLOOP
				}
			}
		}
	}
}

func (buf *File_buf) Keyarrow_procces(high *uint32, width *int, ev termbox.Key, index_buf *Index_buf, search_strs *[]Search_strs) {
	switch ev {
	case termbox.KeyArrowDown:
		*high++
		buf.Drow_Termbox(*high, *width, search_strs, index_buf)
	case termbox.KeyArrowUp:
		if *high == 0 {
			return
		} else {
			*high--
			buf.Drow_Termbox(*high, *width, search_strs, index_buf)
		}
	case termbox.KeyArrowRight:
		*width++
		buf.Drow_Termbox(*high, *width, search_strs, index_buf)
	case termbox.KeyArrowLeft:
		if *width == 0 {
			return
		}
		*width--
		buf.Drow_Termbox(*high, *width, search_strs, index_buf)
	}
	return
}

func (buf *File_buf) Read_File(file string, index_buf *Index_buf) {
	now := time.Now()
	_, high_size := termbox.Size()
	high := uint32(high_size)
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	read := bufio.NewScanner(f)

	var i uint32 = 0
	for ; read.Scan(); i++ {
		(*buf)[i] = read.Text()
		(*index_buf)[i] = Key{i, nil}
		if i == high {
			(*index_buf)[4294967295] = Key{i, nil}
			empty_slice := []Search_strs{{"", 0, 0}}
			buf.Drow_Termbox(0, 0, &empty_slice, index_buf)
		}
	}
	(*index_buf)[4294967295] = Key{i, nil}
	log.Printf("##Read_File##\t%d milisecond\tkey= \"%s\"", time.Since(now).Milliseconds(), "")
}

func (buf *File_buf) Re_Create_buf(high_scroll uint32, width_scroll int, search_strs *[]Search_strs, index_buf *Index_buf) *Index_buf {
	re_now := time.Now()
	var new_index_buf Index_buf = Index_buf{}

	if (*search_strs)[1].str == "s" {
		for key, value := range *index_buf {
			new_index_buf[key] = value
		}
		buf.New_Srot_buf(&new_index_buf)
	} else if (*search_strs)[1].str == "u" {
		buf.New_Uniq_buf(high_scroll, width_scroll, search_strs, index_buf, &new_index_buf)
	} else {
		buf.New_Grep_buf(high_scroll, width_scroll, search_strs, index_buf, &new_index_buf)
	}
	log.Printf("##Re_Create_buf##\t%d milisecond\tkey= \"%s\"", time.Since(re_now).Milliseconds(), (*search_strs)[1].str)
	return &new_index_buf
}

func (buf *File_buf) Drow_Termbox(high_scroll uint32, width_scroll int, search_strs *[]Search_strs, index_buf *Index_buf) {
	now := time.Now()
	termbox.Clear(coldef, coldef)
	width_size, high_size := termbox.Size()
	var sum_search_strs string
	for _, str := range *search_strs {
		sum_search_strs = str.str + " " + sum_search_strs
	}
	sum_search_strs = sum_search_strs[:len(sum_search_strs)-1]
	tbprint(0, 0, sum_search_strs, nil)

	var print_line_nb int = 1
	var index_key uint32 = 0
	var scroll_count uint32 = 0
	tmp, ok := (*index_buf)[4294967295]
	if !ok {
		log.Printf("no such last index(index_buf's key 4294967295)")
	}
	last_index_key := tmp.index

	for print_line_nb < high_size {
		if index_key == last_index_key {
			break
		}
		_, ok := (*index_buf)[index_key]
		if ok == false {
			termbox.Close()
			fmt.Println("no such key")
			os.Exit(0)
		}
		if scroll_count < high_scroll {
			scroll_count++
			index_key++
			continue
		} else {
			tbprint(width_scroll*width_size, print_line_nb, (*buf)[(*index_buf)[index_key].index], (*index_buf)[index_key].match)
			print_line_nb++
			index_key++
		}
	}
	termbox.Flush()
	log.Printf("##Drow_Termbox##\t%d milisecond\tkey= \"%s\"", time.Since(now).Milliseconds(), sum_search_strs)
}

func tbprint(width, y int, str string, color_lenge []int) {
	x := 0
	ch_color := coldef
	bg_color := coldef
	for i := 0; i < len(str); i++ {
		if i < width {
			continue
		}
		if color_lenge == nil {
		} else if color_lenge[0] <= i && i < color_lenge[1] {
			if str[i] == ' ' || str[i] == '\t' {
				bg_color = termbox.ColorLightBlue
				ch_color = coldef
			} else {
				ch_color = termbox.ColorLightBlue
				bg_color = coldef
			}
		} else {
			ch_color = coldef
			bg_color = coldef
		}
		if str[i] == '\t' {
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
		} else {
			termbox.SetCell(x, y, rune(str[i]), ch_color, bg_color)
			x += runewidth.RuneWidth(rune(str[i]))
		}
	}
	if y == 0 {
		termbox.SetCell(x, y, ' ', ch_color, termbox.ColorWhite)
	}
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
