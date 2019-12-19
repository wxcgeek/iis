package action

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/coyove/iis/cmd/ch/config"
	"github.com/coyove/iis/cmd/ch/mv"
	"github.com/gin-gonic/gin"
)

func getAuthor(g *gin.Context) string {
	a := mv.SoftTrunc(g.PostForm("author"), 32)
	//if a == "" {
	//	a = "user/" + hashIP(g)
	//}
	return a
}

func hashIP(g *gin.Context) string {
	ip := append(net.IP{}, g.MustGet("ip").(net.IP)...)
	if len(ip) == net.IPv4len {
		ip[3] = 0 // \24
	} else if len(ip) == net.IPv6len {
		copy(ip[10:], "\x00\x00\x00\x00\x00\x00") // \80
	}
	return ip.String()
}

func New(g *gin.Context) {
	var (
		ip      = hashIP(g)
		content = mv.SoftTrunc(g.PostForm("content"), int(config.Cfg.MaxContent))
		image   = g.PostForm("image64")
		nsfw    = g.PostForm("nsfw") != ""
		redir   = func(a, b string) { g.Redirect(302, "/t"+EncodeQuery(a, b, "content", content)) }
		err     error
	)

	u, ok := g.Get("user")
	if !ok {
		redir("error", "user/not-logged-in")
		return
	}

	if len(content) < 3 && image == "" {
		redir("error", "content/too-short")
		return
	}

	if ret := checkToken(g); ret != "" {
		redir("error", ret)
		return
	}

	if image != "" {
		if image, err = uploadImgur(image); err != nil {
			redir("error", err.Error())
			return
		}
		image = "IMG:" + image
	}

	a := &mv.Article{
		Content: content,
		Media:   image,
		IP:      ip,
		NSFW:    nsfw,
	}

	for i := 1; i <= 8; i++ {
		key := "poll" + strconv.Itoa(i)
		poll := mv.SoftTrunc(g.PostForm(key), 64)
		if poll == "" {
			break
		}
		if a.Extras == nil {
			a.Extras = map[string]string{}
			a.Extras["poll"] = "true"
		}
		a.Extras[key] = poll
	}

	aid, err := m.Post(a, u.(*mv.User))
	if err != nil {
		log.Println(aid, err)
		redir("error", "internal/error")
		return
	}

	content = ""
	redir("", "")
}

func Reply(g *gin.Context) {
	var (
		reply   = g.PostForm("reply")
		ip      = hashIP(g)
		content = mv.SoftTrunc(g.PostForm("content"), int(config.Cfg.MaxContent))
		image   = g.PostForm("image64")
		nsfw    = g.PostForm("nsfw") != ""
		redir   = func(a, b string) { g.Redirect(302, "/p/"+reply+EncodeQuery(a, b, "content", content)) }
		err     error
	)

	u, ok := g.Get("user")
	if !ok {
		redir("error", "user/not-logged-in")
		return
	}

	if ret := checkToken(g); ret != "" {
		redir("error", ret)
		return
	}

	if g.PostForm("delete") != "" && g.PostForm("delete-confirm") != "" {
		u := u.(*mv.User)
		err := m.UpdateArticle(reply, func(a *mv.Article) error {
			if u.ID != a.Author && !u.IsMod() {
				return fmt.Errorf("user/can-not-delete")
			}
			a.Content = mv.DeletionMarker
			a.Media = ""
			return nil
		})
		if err != nil {
			redir("error", err.Error())
		} else {
			redir("", "")
		}
		return
	}

	if g.PostForm("make-nsfw") != "" {
		u := u.(*mv.User)
		err := m.UpdateArticle(reply, func(a *mv.Article) error {
			if u.ID != a.Author && !u.IsMod() {
				return fmt.Errorf("user/can-not-delete")
			}
			a.NSFW = g.PostForm("nsfw-confirm") != ""
			return nil
		})
		if err != nil {
			redir("error", err.Error())
		} else {
			redir("", "")
		}
		return
	}

	if image != "" {
		image, err = uploadImgur(image)
		if err != nil {
			redir("error", err.Error())
			return
		}
		image = "IMG:" + image
	}

	if _, err := m.PostReply(reply, content, image, u.(*mv.User), ip, nsfw); err != nil {
		log.Println(err)
		redir("error", "error/can-not-reply")
		return
	}

	content = ""
	redir("", "")
}
