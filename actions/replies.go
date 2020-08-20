package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
)

func loadReply(c buffalo.Context, id string) (*models.Reply, error) {
	tx := c.Value("tx").(*pop.Connection)
	reply := &models.Reply{}
	if err := c.Bind(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Find(reply, id); err != nil {
		return nil, c.Error(404, err)
	}
	topic := new(models.Topic)
	if err := tx.Find(topic, reply.TopicID); err != nil {
		return nil, c.Error(404, err)
	}
	usr := new(models.User)
	if err := tx.Find(usr, reply.AuthorID); err != nil {
		return nil, c.Error(404, err)
	}
	reply.Topic = topic
	reply.Author = usr
	return reply, nil
}