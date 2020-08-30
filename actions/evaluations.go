package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
)

// CategoriesIndex default implementation.
func EvaluationIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	evals := &models.Evaluations{}
	if err := tx.All(evals); err != nil {
		return c.Error(404, err)
	}
	c.Set("evaluations", evals)
	return c.Render(200, r.HTML("curso/eval-index.plush.html"))
}

func CursoEvaluationCreateGet(c buffalo.Context) error {
	return c.Render(200, r.HTML("curso/eval-create.plush.html"))
}

func CursoEvaluationCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	eval := &models.Evaluation{}
	if err:= c.Bind(eval); err != nil {
		return errors.WithStack(err)
	}
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(eval)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("evaluation", eval)
		c.Flash().Add("success", T.Translate(c,"curso-python-evaluation-add-fail"))
		return c.Render(422, r.HTML("topics/create.plush.html"))
	}
	c.Logger().Info("CursoEvaluationCreatePost success")
	c.Flash().Add("success", T.Translate(c,"curso-python-evaluation-add-success"))
	return c.Render(200, r.HTML("curso/eval-create.plush.html"))
}

func CursoEvaluationGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	eval := &models.Evaluation{}
	eid := c.Param("evalid")
	q := tx.Where("id = ?",eid)
	if err := q.First(eval); err != nil {
		return c.Error(404, err)
	}
	c.Set("evaluation",eval)

	return c.Render(200, r.HTML("curso/eval-get.plush.html"))
}



