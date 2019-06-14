package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/users"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Users paginated fetch.
func Users(c *gin.Context) {
	var (
		limit  = 10
		sort   = c.Query("sort")
		before *bson.ObjectId
		after  *bson.ObjectId
	)
	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n <= 50 {
		limit = n
	}

	if bid := c.Query("before"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		before = &id
	}

	if bid := c.Query("after"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		after = &id
	}

	set, err := user.FetchBy(
		deps.Container,
		user.Page(limit, sort == "reverse", before, after),
	)
	if err != nil {
		panic(err)
	}
	c.JSON(200, set)
}

type upsertBanForm struct {
	RelatedTo string         `json:"related_to" binding:"required,eq=site|eq=post|eq=comment"`
	RelatedID *bson.ObjectId `json:"related_id"`
	UserID    bson.ObjectId  `json:"user_id"`
	Reason    string         `json:"reason" binding:"required"`
	Content   string         `json:"content" binding:"max=255"`
}

// Ban endpoint.
func Ban(c *gin.Context) {
	var form upsertBanForm
	if err := c.BindJSON(&form); err != nil {
		jsonBindErr(c, http.StatusBadRequest, "Invalid ban request, check parameters", err)
		return
	}
	rules := config.C.Rules()
	if _, exists := rules.BanReasons[form.Reason]; false == exists {
		jsonErr(c, http.StatusBadRequest, "Invalid ban category")
		return
	}
	ban, err := users.UpsertBan(deps.Container, users.Ban{
		UserID:    form.UserID,
		RelatedID: form.RelatedID,
		RelatedTo: form.RelatedTo,
		Content:   form.Content,
		Reason:    form.Reason,
	})
	if err != nil {
		panic(err)
	}

	//events.In <- events.NewFlag(flag.ID)
	c.JSON(200, gin.H{"status": "okay", "ban": ban})
}

// BanReasons endpoint.
func BanReasons(c *gin.Context) {
	rules := config.C.Rules()
	reasons := []string{}
	for k := range rules.BanReasons {
		reasons = append(reasons, k)
	}
	c.JSON(200, gin.H{"status": "okay", "reasons": reasons})
}
