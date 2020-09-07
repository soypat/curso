package actions

import (
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/plush/v4"
	"github.com/soypat/curso/models"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("app:templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
			"timeSince":   timeSince,
			"joinPath":    joinPath,
			"displayName": displayName,
			"derefUser":   derefUser,
			"csrf": func() template.HTML {
				return template.HTML("<input name=\"authenticity_token\" value=\"<%= authenticity_token %>\" type=\"hidden\">")
			},
			"codeFmt": codeFmt,
			"codeTheme":            codeTheme,
			"codeThemeFormOptions": codeThemeOptions,
			"avatar": func(user *models.User) template.HTML { // style="height:28px;border-radius:50%;"
				if user.Role == "admin" {
					return template.HTML(fmt.Sprintf(`<img src="%s" img title="%s" alt="%s" class="avatar-img admin">`, user.ImageSrc(), displayName(user), displayName(user)))
				}
				return template.HTML(fmt.Sprintf(`<img src="%s" img title="%s" alt="%s" class="avatar-img">`, user.ImageSrc(), displayName(user), displayName(user)))
			},
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
	return strings.Join(sli, "/") + "/"
}

func displayName(user *models.User) string {
	if user.Nick != "" {
		return user.Nick
	}
	return user.Name
}
func derefUser(u models.User) *models.User { return &u }

func timeSince(created time.Time, ctx plush.HelperContext) string {
	if true && false {
		return created.UTC().Format(time.RFC3339)
	}
	now := time.Now().UTC().Add(-time.Hour * hourDiffUTC)
	delta := now.Sub(created.UTC())
	days := int(math.Abs(delta.Hours()) / 24)
	if days > 30 {
		return created.Format("2006-02-01")
	}
	if days >= 1 {
		return fmt.Sprintf("%dd", days)
	}
	if delta.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(delta.Hours()))
	}
	if delta.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	}
	return fmt.Sprintf("%ds", int(delta.Seconds()))
}

func codeFmt(md, code string, ctx plush.HelperContext) template.HTML {
	md = "```" + code + "\n" + md + "\n```"
	mdHTML, _ := plush.MarkdownHelper(md, ctx)
	return mdHTML
}

const defaultTheme = "idea" //"idea"
const defaultThemeURL = "//cdn.jsdelivr.net/gh/highlightjs/cdn-release@10.1.2/build/styles/idea.min.css"

var codeThemes = map[string]string{
	"default":                  "//cdn.jsdelivr.net/gh/highlightjs/cdn-release@10.1.2/build/styles/default.min.css",
	"android-studio (dark)":    "//cdn.jsdelivr.net/npm/highlight.js@10.1.2/styles/androidstudio.css",
	defaultTheme:               defaultThemeURL,
	"darcula":                  "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/darcula.min.css",
	"github":                   "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/github.min.css",
	"github gist":              "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/github-gist.min.css",
	"kimbie dark":              "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/kimbie.dark.min.css",
	"lightfair":                "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/lightfair.min.css",
	"monokai-sublime (dark)":   "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/monokai-sublime.min.css",
	"escuela":                  "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/school-book.min.css",
	"solarized (dark)":         "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/solarized-dark.min.css",
	"solarized":                "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/solarized-light.min.css",
	"zenburn (dark)":           "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/zenburn.min.css",
	"vs":                       "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/vs.min.css",
	"xt256 (dark)":             "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/xt256.min.css",
	"gradient dark":            "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/gradient-dark.min.css",
	"grayscale":                "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/grayscale.min.css",
	"gml (dark)":               "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/gml.min.css",
	"ir black":                 "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/ir-black.min.css",
	"night owl (dark)":         "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/night-owl.min.css",
	"dark":                     "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/dark.min.css",
	"a 11 y dark":              "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/a11y-dark.min.css",
	"tomorrow night blue":      "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/tomorrow-night-blue.min.css",
	"tomorrow night eighties":  "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/tomorrow-night-eighties.min.css",
	"tomorrow":                 "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/tomorrow.min.css",
	"xcode":                    "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/xcode.min.css",
	"railscasts (dark)":        "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/railscasts.min.css",
	"pojoaque (dark)":          "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/pojoaque.min.css",
	"qt creator dark":          "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/qtcreator_dark.min.css",
	"qt creator light":         "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/qtcreator_light.min.css",
	"shades of purple":         "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/shades-of-purple.min.css",
	"atelier plateau dark":     "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atelier-plateau-dark.min.css",
	"atelier seaside dark":     "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atelier-seaside-dark.min.css",
	"atelier sulphurpool dark": "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atelier-sulphurpool-dark.min.css",
	"atom one dark reasonable": "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atom-one-dark-reasonable.min.css",
	"atom one dark":            "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atom-one-dark.min.css",
	"atom one light":           "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/atom-one-light.min.css",
	"far":                      "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/far.min.css",
	"dracula":                  "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/dracula.min.css",
	"purebasic":                "//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.18.3/styles/purebasic.min.css",
	"satanus":                  "/assets/css/hljs-satanus.css",
}

func codeTheme(theme string) template.HTML {
	url, present := codeThemes[theme]
	if !present {
		url = defaultThemeURL
	}
	return template.HTML(fmt.Sprintf(`<link rel="stylesheet" href="%s">`, url))
}

func codeThemeOptions(u *models.User) template.HTML {
	var form []string
	for name, _ := range codeThemes {
		if name == u.Theme || (u.Theme == "" && name == defaultTheme) {
			form = append(form, fmt.Sprintf("<option name=\"%s\" selected>%s</option>", name, name))
		} else {
			form = append(form, fmt.Sprintf("<option name=\"%s\">%s</option>", name, name))
		}
	}
	sort.Strings(form)
	return template.HTML(strings.Join(form, "\n"))
}
