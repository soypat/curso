package actions

import (
	"encoding/json"
	"errors"
	"github.com/gobuffalo/buffalo"
	"github.com/soypat/curso/models"
	"go.etcd.io/bbolt"
	"strings"
)

var (
	errBucketNotFound = errors.New("did not find " + safeUsersBucketName + " bucket in bolt database")
)

// name must be first created in models/bbolt.go in init()
const safeUsersBucketName = "safeUsers"

type safeUser struct {
	Name  string `json:"nick" gob:"nick"`
	Email string `json:"email" gob:"email"`
	Responsible string `json:"responsible" gob:"resp"`
}

type safeUsers []safeUser

func (s safeUsers) String() (out string) {
	for _,u := range s {
		out+= u.Email + "\n"
	}
	return
}

func SafeListGet(c buffalo.Context) error {
	var users safeUsers
	btx := c.Value("btx").(*bbolt.Tx)
	err := func() error {
		b := btx.Bucket([]byte(safeUsersBucketName))
		if b == nil {
			return errBucketNotFound
		}
		return b.ForEach(func(_, v []byte) error {
			var user safeUser
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})
	}()

	if err != nil {
		return c.Error(500, err)
	}
	c.Set("safe_users",users)
	return c.Render(200, r.HTML("users/safelist.plush.html"))
}

type safeForm struct {
	List string `json:"safelist" form:"safelist"`
}

func SafeListPost(c buffalo.Context) error {
	responsible := c.Value("current_user").(*models.User)
	var form safeForm
	if err:=c.Bind(&form); err!=nil {
		return c.Error(500,err)
	}
	users := safeFormToSafeList(form)
	btx := c.Value("btx").(*bbolt.Tx)
	err := func() error {
		b := btx.Bucket([]byte(safeUsersBucketName))
		if b == nil {
			return errBucketNotFound
		}
		for _,user := range users {
			user.Responsible = responsible.Name
			bson, err:= json.Marshal(user)
			if err!=nil {
				return err
			}
			err = b.Put([]byte(user.Email),bson)
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return c.Error(500,err)
	}
	c.Flash().Add("success","Safelist updated successfully.")
	return c.Redirect(302,"allUsersPath()")
}

type void struct{}

const safeDomain = "itba.edu.ar"


// works kind of like authorize but does not verify user exists.
func SafeList(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		u := c.Value("current_user")
		if u == nil {
			c.Flash().Add("danger", T.Translate(c,"app-user-required"))
			return c.Redirect(302, "/")
		}
		email := u.(*models.User).Email
		two := strings.Split(email, "@")
		if two[1] == safeDomain {
			return next(c)
		}
		var exists bool
		btx := c.Value("btx").(*bbolt.Tx)
		err := btx.DB().View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(safeUsersBucketName))
			if b == nil {
				return errBucketNotFound
			}
			exists = b.Get([]byte(email)) != nil
			return nil
		})
		if err != nil {
			c.Logger().Errorf("SAFELIST MALUFUNCTION: %s",err)
			return next(c)
		}
		if !exists {
			c.Flash().Add("warning", T.Translate(c, "safelist-user-not-found"))
			c.Session().Clear()
			_ = c.Session().Save()
			return c.Redirect(302, "/") //, render.Data{"provider":user.Provider})
		}
		user := u.(*models.User)
		if user.Role == "" {
			user.Role = "safe"
			c.Set("current_user",u)
		}
		return next(c)
	}
}

func safeFormToSafeList(sf safeForm) (SU safeUsers) {
	splits := strings.Split(sf.List,"\n")
	splitComma := strings.Split(sf.List,",")
	splitSColon:= strings.Split(sf.List,";")
	if  len(splitComma) > len(splits) {
		splits = splitComma
	}
	if  len(splitSColon) > len(splits) && len(splitSColon) > len(splitComma) {
		splits = splitSColon
	}
	for _, email := range splits {
		email = strings.TrimSpace(email)
		if !isEmail(email) {
			continue
		}
		SU = append(SU, safeUser{
			Name:  "None",
			Email: email,
		})
	}
	return
}

func isEmail(s string) bool {
	if strings.ContainsAny(s,"\"(),:;<>[\\] \t\n") {
		return false
	}
	dub := strings.Split(s,"@")
	if len(dub) != 2 {
		return false
	}
	back := strings.Split(dub[1],".")
	return len(back) >= 2
}