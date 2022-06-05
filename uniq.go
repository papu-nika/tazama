package main

import (
	"log"
	"time"
)

func (index_buf *IndexBuf) New_Uniq_buf(new_index_buf *IndexBuf) {
	now := time.Now()

	var new_index_key int = 0
	//var print_line_nb int = 1
	var index_key int = 0
	//var scroll_count uint32 = 0
	var match_times int = 1
	//_, high_size := termbox.Size()
	last_index_key := (*index_buf)[0].index

	is_uniq := func(before_str string, str string) bool {
		if before_str == str {
			return true
		} else {
			return false
		}
	}

	for index_key < last_index_key {
		if is_uniq(file_buf[(*index_buf)[index_key].index], file_buf[(*index_buf)[index_key+1].index]) {
			match_times++
			index_key++
		} else {
			(*new_index_buf)[new_index_key] = Keys{
				index:      (*index_buf)[index_key].index,
				uniq_num:   match_times,
				grep_range: nil,
				cut_range:  []int{0, len(file_buf[(*index_buf)[index_key].index])},
			}
			new_index_key++
			index_key++
			match_times = 1
		}
	}
	(*new_index_buf)[0] = Keys{
		index:      new_index_key,
		uniq_num:   0,
		grep_range: nil,
		cut_range:  nil,
	}
	log.Printf("##New_Uniq_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
}
