package main

import (
	"log"
	"time"
)

func (index_buf *IndexBuf) New_Srot_buf(new_index_buf *IndexBuf) {
	now := time.Now()
	if (*index_buf)[0].index < 2 {
		return
	}

	var tmp Keys
	for key, value := range *index_buf {
		if key == 0 {
			tmp = value
			continue
		}
		(*new_index_buf)[key] = value
	}

	pivot := tmp.index - 1
	new_index_buf.QuicSort(1, pivot)
	(*new_index_buf)[0] = tmp
	//log.Printf("%p, new=%p", index_buf, new_index_buf)
	log.Printf("##New_Srot_buf##\t%d milisecond\t", time.Since(now).Milliseconds())
	return
}

func (new_index_buf *IndexBuf) QuicSort(start, last int) {
	var qart int
	log.Println("###", start, last, " = ", last-start)
	if last-start < 15 {
		new_index_buf.InsertSort(start, last)
	} else if start < last {
		qart = new_index_buf.Quicsort_part_left_right(start, last)
		go func(s, l int) {
			new_index_buf.QuicSort(s, l)
		}(start, qart-1)
		go func(s, l int) {
			new_index_buf.QuicSort(s, l)
		}(qart+1, last)
	}
	return
}

func (new_index_buf *IndexBuf) Quicsort_part_left_right(start, last int) int {
	i := start - 1
	pivot := file_buf[(*new_index_buf)[last].index]
	for k := start; k < last; k++ {
		if file_buf[(*new_index_buf)[k].index] < pivot {
			i++
			new_index_buf.Swap_index_buf(i, k)
		}
	}
	new_index_buf.Swap_index_buf(i+1, last)
	return i + 1
}

func (new_index_buf *IndexBuf) InsertSort(start, last int) {

	for ; start < last; start++ {
		for i := 0; i < start; i++ {
			if file_buf[(*new_index_buf)[start-i-1].index] > file_buf[(*new_index_buf)[start-i].index] {
				new_index_buf.Swap_index_buf(start-i-1, start-i)
			} else {
				break
			}
		}
	}
}

func (new_index_buf *IndexBuf) Swap_index_buf(a, b int) {
	(*new_index_buf)[a], (*new_index_buf)[b] = (*new_index_buf)[b], (*new_index_buf)[a]
}
