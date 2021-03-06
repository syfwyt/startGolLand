// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package logic

import (
	"fmt"
	"unicode/utf8"

	"github.com/studygolang/studygolang/model"
)

var (
	publishObservable Observable
	modifyObservable  Observable
	commentObservable Observable
	ViewObservable    Observable
	appendObservable  Observable
	topObservable     Observable
	likeObservable    Observable
)

func init() {
	publishObservable = NewConcreteObservable(actionPublish)
	publishObservable.AddObserver(&UserWeightObserver{})
	publishObservable.AddObserver(&TodayActiveObserver{})
	publishObservable.AddObserver(&UserRichObserver{})

	modifyObservable = NewConcreteObservable(actionModify)
	modifyObservable.AddObserver(&UserWeightObserver{})
	modifyObservable.AddObserver(&TodayActiveObserver{})
	modifyObservable.AddObserver(&UserRichObserver{})

	commentObservable = NewConcreteObservable(actionComment)
	commentObservable.AddObserver(&UserWeightObserver{})
	commentObservable.AddObserver(&TodayActiveObserver{})
	commentObservable.AddObserver(&UserRichObserver{})

	ViewObservable = NewConcreteObservable(actionView)
	ViewObservable.AddObserver(&UserWeightObserver{})
	ViewObservable.AddObserver(&TodayActiveObserver{})
	ViewObservable.AddObserver(&FeedSeqObserver{})

	appendObservable = NewConcreteObservable(actionAppend)
	appendObservable.AddObserver(&UserWeightObserver{})
	appendObservable.AddObserver(&TodayActiveObserver{})
	appendObservable.AddObserver(&UserRichObserver{})

	topObservable = NewConcreteObservable(actionTop)
	topObservable.AddObserver(&UserWeightObserver{})
	topObservable.AddObserver(&TodayActiveObserver{})
	topObservable.AddObserver(&UserRichObserver{})

	likeObservable = NewConcreteObservable(actionLike)
	likeObservable.AddObserver(&UserWeightObserver{})
	likeObservable.AddObserver(&TodayActiveObserver{})
	likeObservable.AddObserver(&UserRichObserver{})
}

type Observer interface {
	Update(action string, uid, objtype, objid int)
}

type Observable interface {
	// AddObserver ???????????????????????????
	AddObserver(o Observer)
	// Detach ???????????????????????????????????????
	RemoveObserver(o Observer)
	// NotifyObservers ?????????????????????????????????
	NotifyObservers(uid, objtype, objid int)
}

const (
	actionPublish = "publish"
	actionModify  = "modify"
	actionComment = "comment"
	actionView    = "view"
	actionAppend  = "append"
	actionTop     = "top"  // ??????
	actionLike    = "like" // ???????????????
)

type ConcreteObservable struct {
	observers []Observer
	action    string
}

func NewConcreteObservable(action string) *ConcreteObservable {
	return &ConcreteObservable{
		action:    action,
		observers: make([]Observer, 0, 8),
	}
}

func (this *ConcreteObservable) AddObserver(o Observer) {
	this.observers = append(this.observers, o)
}

func (this *ConcreteObservable) RemoveObserver(o Observer) {
	if len(this.observers) == 0 {
		return
	}

	var indexToRemove int

	for i, observer := range this.observers {
		if observer == o {
			indexToRemove = i
			break
		}
	}

	this.observers = append(this.observers[:indexToRemove], this.observers[indexToRemove+1:]...)
}

func (this *ConcreteObservable) NotifyObservers(uid, objtype, objid int) {
	for _, observer := range this.observers {
		observer.Update(this.action, uid, objtype, objid)
	}
}

// ///////////////////////// ??????????????? ////////////////////////////////////////

type UserWeightObserver struct{}

func (this *UserWeightObserver) Update(action string, uid, objtype, objid int) {
	if uid == 0 {
		return
	}

	var weight int
	switch action {
	case actionPublish:
		weight = 20
	case actionModify:
		weight = 2
	case actionComment:
		weight = 5
	case actionView:
		weight = 1
	case actionAppend:
		weight = 15
	case actionTop:
		weight = 5
	case actionLike:
		weight = 3
	}

	DefaultUser.IncrUserWeight("uid", uid, weight)
}

type TodayActiveObserver struct{}

func (*TodayActiveObserver) Update(action string, uid, objtype, objid int) {
	if uid == 0 {
		return
	}

	var weight int

	switch action {
	case actionPublish:
		weight = 20
		// ????????????????????????????????????????????????
		incrPublishTimes(uid)
		recordLastPubishTime(uid)
	case actionModify:
		weight = 2
	case actionComment:
		weight = 5
	case actionView:
		weight = 1
	case actionAppend:
		weight = 15
	case actionTop:
		weight = 5
	case actionLike:
		weight = 5
	}

	DefaultRank.GenDAURank(uid, weight)
}

type UserRichObserver struct{}

var objType2MissType = map[int]int{
	model.TypeTopic:    model.MissionTypeTopic,
	model.TypeArticle:  model.MissionTypeArticle,
	model.TypeResource: model.MissionTypeResource,
	model.TypeWiki:     model.MissionTypeWiki,
	model.TypeBook:     model.MissionTypeBook,
	model.TypeProject:  model.MissionTypeProject,
}

// Update ????????????????????? objid ??? cid
func (UserRichObserver) Update(action string, uid, objtype, objid int) {
	if uid == 0 {
		return
	}

	user := DefaultUser.FindOne(nil, "uid", uid)

	var (
		typ   int
		award int
		desc  string
	)

	if action == actionPublish || action == actionComment {
		var comment *model.Comment
		if action == actionComment {
			comment, _ = DefaultComment.FindById(objid)
			if comment.Cid != objid {
				return
			}

			objid = comment.Objid

			award = -5
			typ = model.MissionTypeReply
		} else {
			award = -20
			typ = objType2MissType[objtype]
		}

		switch objtype {
		case model.TypeTopic:
			topic := DefaultTopic.findByTid(objid)
			if topic.Tid != objid {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/topics/%d">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					objid,
					topic.Title)

				if uid != topic.Uid {
					// ???????????????????????????
					replyDesc := fmt.Sprintf(`?????? <a href="/user/%s">%s</a> ????????? ??? <a href="/topics/%d">%s</a>`,
						user.Username,
						user.Username,
						objid,
						topic.Title)
					author := DefaultUser.FindOne(nil, "uid", topic.Uid)
					DefaultUserRich.IncrUserRich(author, model.MissionTypeReplied, 5, replyDesc)
				}
			} else {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/topics/%d">%s</a>`,
					utf8.RuneCountInString(topic.Content),
					objid,
					topic.Title)
			}

		case model.TypeArticle:
			article, err := DefaultArticle.FindById(nil, objid)
			if err != nil {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/articles/%d">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					objid,
					article.Title)
				if article.Domain == WebsiteSetting.Domain && user.Username != article.Author {
					// ???????????????????????????
					replyDesc := fmt.Sprintf(`?????? <a href="/user/%s">%s</a> ????????? ??? <a href="/articles/%d">%s</a>`,
						user.Username,
						user.Username,
						objid,
						article.Title)
					author := DefaultUser.FindOne(nil, "username", article.Author)
					DefaultUserRich.IncrUserRich(author, model.MissionTypeReplied, 5, replyDesc)
				}
			} else {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/articles/%d">%s</a>`,
					utf8.RuneCountInString(article.Txt),
					objid,
					article.Title)
			}
		case model.TypeResource:
			resource := DefaultResource.findById(objid)
			if resource.Id != objid {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/resources/%d">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					objid,
					resource.Title)

				if uid != resource.Uid {
					// ???????????????????????????
					replyDesc := fmt.Sprintf(`?????? <a href="/user/%s">%s</a> ????????? ??? <a href="/resources/%d">%s</a>`,
						user.Username,
						user.Username,
						objid,
						resource.Title)
					author := DefaultUser.FindOne(nil, "uid", resource.Uid)
					DefaultUserRich.IncrUserRich(author, model.MissionTypeReplied, 5, replyDesc)
				}
			} else {

				desc = fmt.Sprintf(`????????????????????? ??? <a href="/resources/%d">%s</a>`,
					objid,
					resource.Title)
			}
		case model.TypeProject:
			project := DefaultProject.FindOne(nil, objid)
			if project == nil || project.Id != objid {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/p/%d">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					objid,
					project.Category+project.Name)

				if user.Username != project.Username {
					// ???????????????????????????
					replyDesc := fmt.Sprintf(`?????? <a href="/user/%s">%s</a> ????????? ??? <a href="/p/%d">%s</a>`,
						user.Username,
						user.Username,
						objid,
						project.Category+project.Name)
					author := DefaultUser.FindOne(nil, "username", project.Username)
					DefaultUserRich.IncrUserRich(author, model.MissionTypeReplied, 5, replyDesc)
				}
			} else {
				desc = fmt.Sprintf(`??????????????????????????? ??? <a href="/p/%d">%s</a>`,
					objid,
					project.Category+project.Name)
			}
		case model.TypeWiki:
			wiki := DefaultWiki.FindById(nil, objid)
			if wiki == nil || wiki.Id != objid {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/wiki/%s">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					wiki.Uri,
					wiki.Title)

				if uid != wiki.Uid {
					// WIKI?????????????????????
					replyDesc := fmt.Sprintf(`?????? <a href="/user/%s">%s</a> ????????? ??? <a href="/wiki/%d">%s</a>`,
						user.Username,
						user.Username,
						objid,
						wiki.Title)
					author := DefaultUser.FindOne(nil, "uid", wiki.Uid)
					DefaultUserRich.IncrUserRich(author, model.MissionTypeReplied, 5, replyDesc)
				}
			} else {
				desc = fmt.Sprintf(`?????????????????? %d ????????????WIKI ??? <a href="/wiki/%s">%s</a>`,
					utf8.RuneCountInString(wiki.Content),
					wiki.Uri,
					wiki.Title)
			}
		case model.TypeBook:
			book, err := DefaultGoBook.FindById(nil, objid)
			if err != nil || book.Id != objid {
				return
			}
			if action == actionComment {
				desc = fmt.Sprintf(`?????????????????? %d ?????????????????? ??? <a href="/book/%d">%s</a>`,
					utf8.RuneCountInString(comment.Content),
					book.Id,
					book.Name)
			} else {
				desc = fmt.Sprintf(`?????????????????? ??? <a href="/book/%d">%s</a>`,
					book.Id,
					book.Name)
			}
		}
	} else if action == actionModify {
		// TODO??????????????????????????????
		// DefaultUserRich.IncrUserRich(uid, model.MissionTypeModify, -2, desc)
		return
	} else if action == actionView {
		return
	} else if action == actionAppend {
		typ = model.MissionTypeAppend
		award = -15
		topic := DefaultTopic.findByTid(objid)
		desc = fmt.Sprintf(`????????? ??? <a href="/topics/%d">%s</a> ????????????`,
			topic.Tid,
			topic.Title)
	} else if action == actionTop {
		typ = model.MissionTypeTop
		award = -200

		switch objtype {
		case model.TypeTopic:
			topic := DefaultTopic.findByTid(objid)
			desc = fmt.Sprintf(`????????? ??? <a href="/topics/%d">%s</a> ??????`,
				topic.Tid,
				topic.Title)
		case model.TypeArticle:
			article, _ := DefaultArticle.FindById(nil, objid)
			desc = fmt.Sprintf(`????????? ??? <a href="/articles/%d">%s</a> ??????`,
				article.Id,
				article.Title)
		}
	} else if action == actionLike {
		// TODO: ???????????????
		return
	}

	DefaultUserRich.IncrUserRich(user, typ, award, desc)
}

type FeedSeqObserver struct{}

func (this *FeedSeqObserver) Update(action string, uid, objtype, objid int) {
	if objid == 0 {
		return
	}

	if action == actionView {
		DefaultFeed.updateSeq(objid, objtype, 0, 0, 1)
	}
}
