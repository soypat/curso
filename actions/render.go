package actions

import (
    "fmt"
    "github.com/gobuffalo/packr/v2"
  "github.com/gobuffalo/buffalo/render"
    "github.com/gobuffalo/plush"
    "github.com/soypat/curso/models"
    "math"
    "strings"
    "time"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

func init() {
  r = render.New(render.Options{
      // HTML layout to be used for all HTML requests:
      HTMLLayout:     "application.plush.html",

      // Box containing all of the templates:
      TemplatesBox: packr.New("app:templates", "../templates"),
      AssetsBox:    assetsBox,

      // Add template helpers here:
      Helpers: render.Helpers{
        // for non-bootstrap form helpers uncomment the lines
        // below and import "github.com/gobuffalo/helpers/forms"
        // forms.FormKey:     forms.Form,
        // forms.FormForKey:  forms.FormFor,
          "timeSince": timeSince,
          "joinPath": joinPath,
          "displayName": displayName,
      },
  })
}
func joinPath(sli ...string) string {
    for i, s := range sli {
        s = strings.TrimSuffix(s, "/")
        if i > 0 {
            s = strings.TrimPrefix(s, "/")
        }
        sli[i] = s
    }
    return strings.Join(sli,"/") +"/"
}

func displayName(user models.User) string {
    if user.Nick != "" {
        return user.Nick
    }
    return user.Name
}

func timeSince(created time.Time, ctx plush.HelperContext) string {
    if true && false {
        return created.UTC().Format(time.RFC3339)
    }
    now := time.Now().UTC()
    delta := now.Sub(created.UTC())
    days := int(math.Abs(delta.Hours()) / 24)
    if days > 30 {
        return created.Format("2006-02-01")
    }
    if days >= 1 {
        return fmt.Sprintf("%dj", days)
    }
    if delta.Hours() >= 1 {
        return fmt.Sprintf("%dh", int(delta.Hours()))
    }
    if delta.Minutes() >= 1 {
        return fmt.Sprintf("%dm", int(delta.Minutes()))
    }
    return fmt.Sprintf("%ds", int(delta.Seconds()))
}
