// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package model

import (
	"fmt"
	"regexp"
	"strings"
)

// 文档对象（供solr使用）
type Document struct {
	Id      string `json:"id"`
	Objid   int    `json:"objid"`
	Objtype int    `json:"objtype"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	PubTime string `json:"pub_time"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
	Viewnum int    `json:"viewnum"`
	Cmtnum  int    `json:"cmtnum"`
	Likenum int    `json:"likenum"`

	HlTitle   string `json:",omitempty"` // 高亮的标题
	HlContent string `json:",omitempty"` // 高亮的内容
}

func NewDocument(object interface{}, objectExt interface{}) *Document {
	var document *Document
	switch objdoc := object.(type) {
	case *Topic:
	case *Article:
		document = &Document{
			Id:      fmt.Sprintf("%d%d", TYPE_ARTICLE, objdoc.Id),
			Objid:   objdoc.Id,
			Objtype: TYPE_ARTICLE,
			Title:   filterTxt(objdoc.Title),
			Author:  objdoc.AuthorTxt,
			PubTime: objdoc.PubDate,
			Content: filterTxt(objdoc.Txt),
			Tags:    objdoc.Tags,
			Viewnum: objdoc.Viewnum,
			Cmtnum:  objdoc.Cmtnum,
			Likenum: objdoc.Likenum,
		}
	case *Resource:
	case *Wiki:
	}

	return document
}

var docRe = regexp.MustCompile("[\r　\n  \t\v]+")
var docSpaceRe = regexp.MustCompile("[ ]+")

// 文本过滤（预处理）
func filterTxt(txt string) string {
	txt = strings.TrimSpace(strings.TrimPrefix(txt, "原"))
	txt = strings.TrimSpace(strings.TrimPrefix(txt, "荐"))
	txt = strings.TrimSpace(strings.TrimPrefix(txt, "顶"))
	txt = strings.TrimSpace(strings.TrimPrefix(txt, "转"))

	txt = docRe.ReplaceAllLiteralString(txt, " ")
	return docSpaceRe.ReplaceAllLiteralString(txt, " ")
}

type AddCommand struct {
	Doc          *Document `json:"doc"`
	Boost        float64   `json:"boost,omitempty"`
	Overwrite    bool      `json:"overwrite"`
	CommitWithin int       `json:"commitWithin,omitempty"`
}

func NewDefaultArgsAddCommand(doc *Document) *AddCommand {
	return NewAddCommand(doc, 0.0, true, 0)
}

func NewAddCommand(doc *Document, boost float64, overwrite bool, commitWithin int) *AddCommand {
	return &AddCommand{
		Doc:          doc,
		Boost:        boost,
		Overwrite:    overwrite,
		CommitWithin: commitWithin,
	}
}

type ResponseBody struct {
	NumFound int         `json:"numFound"`
	Start    int         `json:"start"`
	Docs     []*Document `json:"docs"`
}

type Highlighting struct {
	Title   []string `json:"title"`
	Content []string `json:"content"`
}

type SearchResponse struct {
	RespHeader map[string]interface{}   `json:"responseHeader"`
	RespBody   *ResponseBody            `json:"response"`
	Highlight  map[string]*Highlighting `json:"highlighting"`
}