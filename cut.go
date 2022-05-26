package main

func (index_buf *IndexBuf) New_Cut_buf(new_index_buf *IndexBuf, cut int) {
	var i int
	for i = 1; (*index_buf)[0].index > i; i++ {
		start := 0
		end := 0
		cut_count := 0
		for {
			if start > len([]rune(file_buf[(*index_buf)[i].index])) {
				end = 0
				break
			}
			if cut == cut_count {
				rune_tmp := []rune(file_buf[(*index_buf)[i].index])[start:]
				end = rune_index(rune_tmp, ' ') + start
				//log.Println("start=", start, " end=", end, "cut_count", cut_count)
				break
			}
			rune_tmp := []rune(file_buf[(*index_buf)[i].index])
			//log.Println("start=", start, " end=", end, "cut_count", cut_count)
			start = rune_index(rune_tmp[start:], ' ') + start + 1
			cut_count++
		}
		//log.Println(cut_range, "####", (*index_buf)[i].index)
		(*new_index_buf)[i] = Keys{
			index:      (*index_buf)[i].index,
			uniq_num:   0,
			grep_range: nil,
			cut_range:  []int{start, end},
		}
	}
	(*new_index_buf)[0] = Keys{
		index:      i,
		uniq_num:   0,
		grep_range: nil,
		cut_range:  nil,
	}
}

func rune_index(str []rune, s rune) int {
	var i int
	for i = 0; i < len(str); i++ {
		if str[i] == s {
			break
		}
	}
	return i
}
