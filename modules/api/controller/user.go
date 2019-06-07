package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/flags"
	"github.com/tryanzu/core/board/users"
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
	RelatedTo string        `json:"related_to" binding:"required,eq=site|eq=post|eq=comment"`
	RelatedID bson.ObjectId `json:"related_id"`
	Category  string        `json:"category" binding:"required"`
	Content   string        `json:"content" binding:"max=255"`
}

// Ban endpoint.
func Ban(c *gin.Context) {
	var form upsertFlagForm
	if err := c.BindJSON(&form); err != nil {
		jsonBindErr(c, http.StatusBadRequest, "Invalid ban request, check parameters", err)
		return
	}
	category, err := users.CastCategory(form.Category)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Invalid ban category")
		return
	}
	usr := c.MustGet("user").(user.User)
	if count := flags.TodaysCountByUser(deps.Container, usr.Id); count > 10 {
		jsonErr(c, http.StatusPreconditionFailed, "Can't flag anymore for today")
		return
	}
	ban, err := users.UpsertBan(deps.Container, users.Ban{
		UserID:    usr.Id,
		RelatedID: form.RelatedID,
		RelatedTo: form.RelatedTo,
		Content:   form.Content,
		Category:  category,
	})

	if err != nil {
		jsonErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	//events.In <- events.NewFlag(flag.ID)
	c.JSON(200, ban)
}