package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func (buf *File_buf) Drow_Termbox(high_scroll uint32, width_scroll int, search_strs *[]Search_strs, index_buf *Index_buf, error_message string) {
	now := time.Now()
	termbox.Clear(coldef, coldef)
	width_size, high_size := termbox.Size()
	var sum_search_strs string
	for _, str := range *search_strs {
		sum_search_strs = str.str + " " + sum_search_strs
	}
	sum_search_strs = sum_search_strs[:len(sum_search_strs)-1]
	prompt_print(0, 0, sum_search_strs)

	var print_line_nb int = 1
	var index_key uint32 = 0
	var scroll_count uint32 = 0
	tmp, ok := (*index_buf)[4294967295]
	if !ok {
		log.Printf("no such last index(index_buf's key 4294967295)")
	}
	last_index_key := tmp.index

	for print_line_nb < high_size-1 {
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
			if len((*search_strs)) < 2 {
				index_buf.Tbprint(width_scroll*width_size, 0, print_line_nb, (*buf)[(*index_buf)[index_key].index], index_key)
			} else if (*search_strs)[1].str == "uniq" {
				index_buf.Tbprint(width_scroll*width_size, 0, print_line_nb, (*buf)[(*index_buf)[index_key].index], index_key)
			} else if (*search_strs)[1].str == "uniq-c" {
				num := strconv.Itoa(int((*index_buf)[index_key].uniq))
				var x int
				for x, ch := range num {
					termbox.SetCell(x, print_line_nb, rune(ch), termbox.ColorLightMagenta, coldef)
					x += runewidth.RuneWidth(rune(ch))
				}

				index_buf.Tbprint(width_scroll*width_size, x+2, print_line_nb, (*buf)[(*index_buf)[index_key].index], index_key)
			} else {
				index_buf.Tbprint(width_scroll*width_size, 0, print_line_nb, (*buf)[(*index_buf)[index_key].index], index_key)
			}
			print_line_nb++
			index_key++
		}
	}
	if error_message != "" {
		index_buf.Tbprint(0, 0, print_line_nb, error_message, 0)
	}
	termbox.Flush()
	log.Printf("##Drow_Termbox##\t%d milisecond\tkey= \"%s\"", time.Since(now).Milliseconds(), sum_search_strs)
}

func (index_buf *Index_buf) Tbprint(width, x, y int, str string, index_key uint32) {
	color_lenge := (*index_buf)[index_key].match
	print_renge := (*index_buf)[index_key].cut
	ch_color := coldef
	bg_color := coldef
	str_rune := []rune(str)
	print_renge[1] = len(str_rune)

	for i := print_renge[0]; i < print_renge[1]; i++ {
		if i < width {
			continue
		}
		if color_lenge == nil {
		} else if color_lenge[0] <= i && i < color_lenge[1] {
			if str_rune[i] == ' ' || str_rune[i] == '\t' {
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
		if str_rune[i] == '\t' {
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
			termbox.SetCell(x, y, ' ', ch_color, bg_color)
			x++
		} else {
			termbox.SetCell(x, y, rune(str_rune[i]), ch_color, bg_color)
			x += runewidth.RuneWidth(rune(str_rune[i]))
		}
	}
	if y == 0 {
		termbox.SetCell(x, y, ' ', ch_color, termbox.ColorWhite)
	}
}

func prompt_print(width, y int, str string) {
	x := 0
	ch_color := coldef
	bg_color := coldef

	for i := 0; i < len(str); i++ {
		if i < width {
			continue
		}
		// if color_lenge == nil {
		// } else if color_lenge[0] <= i && i < color_lenge[1] {
		// 	if str[i] == ' ' || str[i] == '\t' {
		// 		bg_color = termbox.ColorLightBlue
		// 		ch_color = coldef
		// 	} else {
		// 		ch_color = termbox.ColorLightBlue
		// 		bg_color = coldef
		// 	}
		// } else {
		// 	ch_color = coldef
		// 	bg_color = coldef
		// }
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

func error_print(str string) {
	x := 0
	ch_color := coldef
	bg_color := coldef
	_, y := termbox.Size()
	y--
	for i := 0; i < len(str); i++ {
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
	termbox.Flush()
}
