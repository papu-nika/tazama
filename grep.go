package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

func (index_buf *IndexBuf) New_Grep_buf(new_index_buf *IndexBuf, search_str string) {
	now := time.Now()
	var new_index_key int = 1
	var print_line_nb int = 1
	var index_key int = 1

	var is_match_index []int
	_, high_size := termbox.Size()
	last_index_key := (*index_buf)[0].index

	var search_func func(str string) []int
	find, err := regexp.CompilePOSIX(search_str)
	var is_meta_ch bool = false
	for _, v := range search_str {
		if !('a' < v && v < 'z' || 'Z' < v && v < 'Z' || '0' < v && v < '9') {
			is_meta_ch = true
		}
	}
	if is_meta_ch {
		search_func = func(str string) []int {
			return find.FindStringIndex(str)
		}
	} else {
		search_func = func(str string) []int {
			is_start_index := strings.Index(str, search_str)
			if is_start_index == -1 {
				return nil
			} else {
				return []int{is_start_index, is_start_index + len(search_str)}
			}
		}
	}

	for index_key < last_index_key {
		if err != nil {
			error_proccess(err)
		}
		is_match_index = search_func(file_buf[(*index_buf)[index_key].index])
		if is_match_index == nil {
			index_key++
			continue
		} else {
			if print_line_nb == high_size {
				(*new_index_buf)[0] = Keys{
					index:      new_index_key,
					uniq_num:   0,
					grep_range: nil,
					cut_range:  nil,
				}
				index_buf.Draw_Termbox()
				termbox.Flush()
			}
			(*new_index_buf)[new_index_key] = Keys{
				index:      (*index_buf)[index_key].index,
				uniq_num:   0,
				grep_range: is_match_index,
				cut_range:  []int{0, len(file_buf[(*index_buf)[index_key].index])},
			}
			print_line_nb++
			index_key++
			new_index_key++
		}
	}
	(*new_index_buf)[0] = Keys{new_index_key, 0, nil, nil}
	//log.Printf("%p, new=%p", index_buf, new_index_buf)
	log.Printf("##New_Grep_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
}
