package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
	"sort"
)

func TopicGet(c buffalo.Context) error {
	f:=c.Value("forum").(*models.Forum)
	tid:=c.Param("tid")
	renderData := render.Data{"forum_title":f.Title,"cat_title":c.Param("cat_title"), "tid":tid}
	topic, err:=loadTopic(c, tid)
	if err != nil {
		return c.Redirect(302,"catPath()", renderData)
	}
	return c.Render(200, "")

	return c.Redirect(200,"topicGetPath()", renderData)//c.Render(200,r.HTML("topics/index.plush.html"))
}

func TopicCreateGet(c buffalo.Context) error {
	return c.Render(200,r.HTML("topics/create.plush.html"))
}

func TopicCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	topic := &models.Topic{}
	if err := c.Bind(topic); err != nil {
		return errors.WithStack(err)
	}
	topic.Author = c.Value("current_user").(*models.User)
	cat := new(models.Category)
	q  := tx.Where("title = ?", c.Param("cat_title"))
	err := q.First(cat)
	if err != nil {
		c.Flash().Add("danger","Error while seeking category")
		return c.Redirect(302,"forumPath()")
	}
	topic.Category = cat
	topic.AuthorID = topic.Author.ID
	topic.CategoryID = topic.Category.ID
	topic.AddSubscriber(topic.AuthorID)
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(topic)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("topic", topic)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("topics/create.plush.html"))
	}
	//err = newTopicNotify(c, topic)
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	f := c.Value("forum").(*models.Forum)
	c.Logger().Info("TopicCreatePost finished success")
	c.Flash().Add("success", "New topic added successfully.")
	return c.Redirect(302, "fCPath()",render.Data{"forum_title":f.Title,"cat_title":cat.Title})
	//return c.Render(200,r.HTML("topics/create.plush.html"))
}





func loadTopic(c buffalo.Context, tid string) (*models.Topic, error) {
	tx := c.Value("tx").(*pop.Connection)
	topic := &models.Topic{}
	if err := c.Bind(topic); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Find(topic, tid); err != nil {
		return nil, c.Error(404, err)
	}
	cat := new(models.Category)
	if err := tx.Find(cat, topic.CategoryID); err != nil {
		return nil, c.Error(404, err)
	}
	usr := new(models.User)
	if err := tx.Find(usr, topic.AuthorID); err != nil {
		return nil, c.Error(404, err)
	}
	if err := tx.BelongsTo(topic).All(&topic.Replies); err != nil {
		return nil, c.Error(404, err)
	}
	topic.Category = cat
	topic.Author = usr
	replies := make(models.Replies, 0, len(topic.Replies))
	for i := range topic.Replies {
		reply, err := loadReply(c, topic.Replies[i].ID.String())
		if err != nil {
			return nil, c.Error(404, err)
		}
		if reply.Deleted {
			continue
		}
		replies = append(replies, *reply)
	}
	topic.Replies = replies
	sort.Sort(topic.Replies)
	return topic, nil
}
