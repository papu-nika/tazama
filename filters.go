package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

func (buf *File_buf) New_Grep_buf(high_scroll uint32, width_scroll int, search_strs *[]Search_strs, index_buf *Index_buf, new_index_buf *Index_buf) {
	now := time.Now()
	var new_index_key uint32 = 0
	var print_line_nb int = 1
	var index_key uint32 = 0
	var scroll_count uint32 = 0
	var is_match_index []int
	_, high_size := termbox.Size()
	last_index_key := (*index_buf)[4294967295].index

	var search_func func(str string) []int
	find, err := regexp.CompilePOSIX((*search_strs)[1].str)
	var is_meta_ch bool = false
	for _, v := range (*search_strs)[1].str {
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
			is_start_index := strings.Index(str, (*search_strs)[1].str)
			if is_start_index == -1 {
				return nil
			} else {
				return []int{is_start_index, is_start_index + len((*search_strs)[1].str)}
			}
		}
	}

	for index_key < last_index_key {
		if err != nil {
			error_proccess(err)
		}
		is_match_index = search_func((*buf)[(*index_buf)[index_key].index])
		if is_match_index == nil {
			index_key++
			continue
		} else {
			if scroll_count < high_scroll {
				scroll_count++
				(*new_index_buf)[new_index_key] = (*index_buf)[index_key]
				index_key++
				new_index_key++
				continue
			} else {
				if print_line_nb == high_size {
					(*new_index_buf)[4294967295] = Key{
						index: new_index_key,
						uniq:  0,
						match: nil,
						cut:   nil,
					}
					buf.Drow_Termbox(high_scroll, width_scroll, search_strs, new_index_buf, "")
				}
				(*new_index_buf)[new_index_key] = Key{
					index: (*index_buf)[index_key].index,
					uniq:  0,
					match: is_match_index,
					cut:   []int{0, len((*buf)[(*index_buf)[index_key].index])},
				}

				print_line_nb++
				index_key++
				new_index_key++
			}
		}
	}
	(*new_index_buf)[4294967295] = Key{new_index_key, 0, nil, nil}
	log.Printf("##New_Grep_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
}

func (buf *File_buf) New_Uniq_buf(high_scroll uint32, width_scroll int, search_strs *[]Search_strs, index_buf *Index_buf, new_index_buf *Index_buf) {
	now := time.Now()

	var new_index_key uint32 = 0
	//var print_line_nb int = 1
	var index_key uint32 = 0
	//var scroll_count uint32 = 0
	var match_times uint32 = 1
	//_, high_size := termbox.Size()
	last_index_key := (*index_buf)[4294967295].index

	is_uniq := func(before_str string, str string) bool {
		if before_str == str {
			return true
		} else {
			return false
		}
	}

	log.Print((*buf)[(*index_buf)[index_key].index], " ", (*buf)[(*index_buf)[index_key+1].index], " ", (*buf)[(*index_buf)[index_key].index+2], " ", (*buf)[(*index_buf)[index_key].index+3], " ", (*buf)[(*index_buf)[index_key].index+4], " ", (*buf)[(*index_buf)[index_key].index+5], " ", (*buf)[(*index_buf)[index_key].index+6], " ", (*buf)[(*index_buf)[index_key].index+7], " ", (*buf)[(*index_buf)[index_key].index+8])

	for index_key < last_index_key {
		if is_uniq((*buf)[(*index_buf)[index_key].index], (*buf)[(*index_buf)[index_key+1].index]) {
			match_times++
			index_key++
		} else {
			(*new_index_buf)[new_index_key] = Key{
				index: (*index_buf)[index_key].index,
				uniq:  match_times,
				match: nil,
				cut:   []int{0, len((*buf)[(*index_buf)[index_key].index])},
			}
			new_index_key++
			index_key++
			match_times = 1
		}
	}
	(*new_index_buf)[4294967295] = Key{
		index: new_index_key,
		uniq:  0,
		match: nil,
		cut:   nil,
	}
	log.Printf("##New_Uniq_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
}

func (buf *File_buf) New_Srot_buf(index_buf *Index_buf) {
	now := time.Now()
	if (*index_buf)[4294967295].index < 2 {
		return
	}
	var new_index_buf Index_buf = *index_buf
	pivot := int64((*index_buf)[4294967295].index) - 1
	buf.QuicSort(&new_index_buf, 0, pivot)
	log.Printf("##New_Srot_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
}

func (buf *File_buf) QuicSort(index_buf *Index_buf, start, last int64) {
	var qart int64
	if start < last {
		qart = buf.Quicsort_part_left_right(index_buf, start, last)
		buf.QuicSort(index_buf, start, qart-1)
		buf.QuicSort(index_buf, qart+1, last)
	}
	return
}

func (buf *File_buf) Quicsort_part_left_right(index_buf *Index_buf, start, last int64) int64 {
	i := start - 1
	pivot := (*buf)[(*index_buf)[uint32(last)].index]
	for k := start; k < last; k++ {
		if (*buf)[(*index_buf)[uint32(k)].index] < pivot {
			i++
			index_buf.Swap_index_buf(i, k)
		}
	}
	index_buf.Swap_index_buf(i+1, last)
	return i + 1
}

func (index_buf *Index_buf) Swap_index_buf(a, b int64) {
	tmp := (*index_buf)[uint32(a)]
	(*index_buf)[uint32(a)] = (*index_buf)[uint32(b)]
	(*index_buf)[uint32(b)] = tmp
}
