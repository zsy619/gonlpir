/***************************************************************
 *
 * Copyright (c) 2015, Menglong TAN <tanmenglong@gmail.com>
 *
 * This program is free software; you can redistribute it
 * and/or modify it under the terms of the GPL licence
 *
 **************************************************************/

/**
 *
 *
 * @file gonlpir.go
 * @author Menglong TAN <tanmenglong@gmail.com>
 * @date Sat Mar  7 13:36:13 2015
 *
 **/

package gonlpir

/*
#include <stdio.h>
#include <stdlib.h>
#include <NLPIR.h>

#cgo linux CFLAGS: -I./include -DOS_LINUX
#cgo linux LDFLAGS: -L./lib/linux64 -lNLPIR
*/
import "C"
import "unsafe"

import (
	"fmt"
)

const (
	GBK        = iota // GBK简体中文
	UTF8              // UTF8简体中文
	BIG5              // 台湾大5码
	GBK_FANTI         // GBK繁体中文
	UTF8_FANTI        // UTF8繁体中文
)

type NLPIR struct {
	init     bool
	dataPath string
	encoding int
	licence  string
}

type Result struct {
	word     string
	spos     string
	ipos     int
	wordId   int
	wordType int
	weight   int
}

func NewResult() *Result {
	p := &Result{}
	return p
}

func NewNLPIR(dataPath string, encoding int, licence string) (*NLPIR, error) {
	p := &NLPIR{}

	p.dataPath = dataPath
	p.encoding = encoding
	p.licence = licence

	d := C.CString(dataPath)
	defer C.free(unsafe.Pointer(d))

	l := C.CString(licence)
	defer C.free(unsafe.Pointer(l))

	if ret := int(C.NLPIR_Init(d, C.int(encoding), l)); ret == 0 {
		return nil, fmt.Errorf("init failed")
	}

	global_init = true

	return p, nil
}

var global_init bool = false

func (this *NLPIR) Exit() {
	if global_init {
		C.NLPIR_Exit()
	}
	global_init = false
}

func (this *NLPIR) ParagraphProcess(paragraph string, needPosTagged bool) string {
	cs := C.CString(paragraph)
	defer C.free(unsafe.Pointer(cs))
	r := C.NLPIR_ParagraphProcess(cs, C.int(BoolToInt(needPosTagged)))
	return C.GoString(r)
}

func (this *NLPIR) ParagraphProcessA(paragraph string, useUserDict bool) []*Result {
	cs := C.CString(paragraph)
	defer C.free(unsafe.Pointer(cs))
	n := C.int(0)

	p := C.NLPIR_ParagraphProcessA(cs, &n, C.int(BoolToInt(useUserDict)))
	C.free(unsafe.Pointer(p))

	r := (*[1 << 30]C.struct_Result)(unsafe.Pointer(p))[:n:n]

	b := []byte(paragraph)
	results := []*Result{}

	for i := 0; i < len(r); i++ {
		res := NewResult()
		res.word = string(b[r[i].start : r[i].start+r[i].length])
		for j := 0; j < len(r[i].sPOS); j++ {
			if r[i].sPOS[j] == 0 {
				continue
			}
			res.spos += string(r[i].sPOS[j])
		}
		res.ipos = int(r[i].iPOS)
		res.wordId = int(r[i].word_ID)
		res.wordType = int(r[i].word_type)
		res.weight = int(r[i].weight)

		results = append(results, res)
	}

	return results
}

func (this *NLPIR) ImportUserDict(fileName string, isOverwrite bool) {
	cs := C.CString(fileName)
	defer C.free(unsafe.Pointer(cs))

	C.NLPIR_ImportUserDict(cs, C.int(BoolToInt(isOverwrite)))
}
