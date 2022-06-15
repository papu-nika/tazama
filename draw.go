package main

import (
	"log"
	"strconv"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func textAreaClear(fgColor termbox.Attribute, bgColor termbox.Attribute) {
	widthSize, highSize := termbox.Size()
	var brank rune = ' '

	for high := 1; high < highSize; high++ {
		for width := 0; width < widthSize; width++ {
			termbox.SetCell(width, high, brank, fgColor, bgColor)
		}
	}

}

func (index_buf *IndexBuf) Draw_Termbox() {
	var search_strs []string = []string{"tets"}
	//termbox.Clear(default_color, default_color)
	textAreaClear(default_color, default_color)

	widthSize, highSize := termbox.Size()
	var term_line_nb int = 1
	var index_key int = scroll.high + 1

	tmp, ok := (*index_buf)[0]
	if !ok {
		log.Printf("no such last index(index_buf's key 0)")
	}
	last_index_key := tmp.index

	for term_line_nb < highSize-1 && index_key != last_index_key {
		if !index_buf.Check_key(index_key) {
			return
		}

		if len(search_strs) > 1 {
			if search_strs[1] == "uniq-c" {
				num := strconv.Itoa(int((*index_buf)[index_key].uniq_num))
				var x int
				for x, ch := range num {
					termbox.SetCell(x, term_line_nb, rune(ch), termbox.ColorLightMagenta, default_color)
					x += runewidth.RuneWidth(rune(ch))
				}
				index_buf.PrintLine(term_line_nb, x+2, widthSize, file_buf[(*index_buf)[index_key].index], index_key)
			} else {
				str := file_buf[(*index_buf)[index_key].index]
				index_buf.PrintLine(term_line_nb, 0, widthSize, str[scroll.width*widthSize:], index_key)
			}
		} else {
			str := file_buf[(*index_buf)[index_key].index]
			index_buf.PrintLine(term_line_nb, 0, widthSize, str[scroll.width*widthSize:], index_key)
		}
		term_line_nb++
		index_key++
	}
	termbox.Flush()
}

func (index_buf *IndexBuf) PrintLine(y, x, widthSize int, str string, index_key int) {
	color_lenge := (*index_buf)[index_key].grep_range
	print_renge := (*index_buf)[index_key].cut_range
	ch_color := default_color
	bg_color := default_color
	str_rune := []rune(str)
	//log.Print((*index_buf)[index_key].cut_range)
	//print_renge[1] = len(str_rune)

	//log.Println(print_renge, "printline")

	for i := print_renge[0]; i < print_renge[1]; i++ {
		if color_lenge == nil {
		} else if color_lenge[0] <= i && i < color_lenge[1] {
			ch_color = termbox.ColorLightBlue
			bg_color = default_color
		} else {
			ch_color = default_color
			bg_color = default_color
		}
		if str_rune[i] == '\t' {
			for i := 0; i < 4; i++ {
				termbox.SetCell(x, y, ' ', ch_color, bg_color)
				x++
			}
		} else {
			termbox.SetCell(x, y, rune(str_rune[i]), ch_color, bg_color)
			x += runewidth.RuneWidth(rune(str_rune[i]))
		}
	}
	if y == 0 {
		termbox.SetCell(x, y, ' ', ch_color, termbox.ColorWhite)
	}
}

func (promptStrs *PromptStrs) PromptPrint() {
	// log.Println("Prompt prrint")
	var prompt rune = '>'
	var str string

	for _, v := range *promptStrs {
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

	widthSize, _ := termbox.Size()
	for x++; x < widthSize; x++ {
		termbox.SetCell(x, 0, ' ', default_color, default_color)
	}
	termbox.Flush()
}

func error_print(error_message string) {
	_, high := termbox.Size()
	error_message_rune := []rune(error_message)

	x := 0
	for i := 0; i < len(error_message_rune); i++ {
		termbox.SetCell(x, high-1, rune(error_message_rune[i]), warning_color, default_color)
		x += runewidth.RuneWidth(rune(error_message_rune[i]))
	}
	termbox.SetCell(x, high-1, ' ', default_color, termbox.ColorWhite)
}
