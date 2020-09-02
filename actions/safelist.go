package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/soypat/curso/models"
	"strings"
)

type void struct{}
var null = void{}

const safeDomain = "itba.edu.ar"
// nasty set implementation
var safelist = map[string]void {
	"pwhittingslow@itba.edu.a":null,
	"mbergerman@itba.edu.ar":null,
	"learodriguez@itba.edu.ar":null,
	"anowik@itba.edu.ar":null,
	"fledesma@itba.edu.ar":null,
	"poseroff@itba.edu.ar":null,
	"fbasili@itba.edu.ar":null,
	"pgarcia@itba.edu.ar":null,
}

func SafeList(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		u:=c.Value("current_user")
		if u == nil {
			return next(c)
		}
		email := u.(*models.User).Email
		two := strings.Split(email,"@")
		if two[1] == safeDomain {
			return next(c)
		}
		_, present := safelist[email]
		if !present {
			c.Flash().Add("warning", T.Translate(c,"safelist-user-not-found"))
			c.Session().Clear()
			_=c.Session().Save()
			return c.Redirect(302,"/") //, render.Data{"provider":user.Provider})
		}
		return next(c)
	}
}
