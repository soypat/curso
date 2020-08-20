package actions

import (
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/buffalo"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/markbates/goth/gothic"
	"github.com/unrolled/secure"

	"github.com/soypat/curso/models"
	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	csrf "github.com/gobuffalo/mw-csrf"
	i18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
// GO_END = production for deployment
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_curso_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))
		// Setup and use translations:
		app.Use(translations())

		// -- Authorization/Security procedures --
		// sets user data in context from session data.
		app.Use(SetCurrentUser)
		app.Use(SiteStruct)
		bah := buffalo.WrapHandlerFunc(gothic.BeginAuthHandler) // Begin authorization handler = bah
		auth := app.Group("/auth")
		auth.GET("/",AuthHome)
		auth.GET("/{provider}/callback", AuthCallback)
		auth.GET("/{provider}", bah)
		auth.Middleware.Skip(Authorize, bah, AuthCallback) // don't ask for authorization on authorization page
		//auth.Middleware.Skip(SetCurrentUser,bah, AuthCallback) // set current user needs to seek user in db. if no users present in db setcurrentuser fails
		auth.DELETE("", AuthDestroy)

		// TODO add authorization and admin auth
		// home page setup
		app.GET("/", manageForum)
		//app.Use(SetCurrentForum)
		app.GET("/f", NotFound)
		forum := app.Group("/f/{forum_title}")
		forum.Use(SetCurrentForum)
		forum.GET("/", forumIndex).Name("forum")
		forum.GET("/create",CategoriesCreateGet).Name("catCreate")
		forum.POST("/create",CategoriesCreatePost)
		//forum.GET("/c/{cat_title}/", CategoriesIndex)
		//forum.GET("/c/{cat_title}/new", TopicCreateGet )
		//forum.POST("/c/{cat_title}/new", TopicCreatePost )
		catGroup := forum.Group("/c/{cat_title}")
		catGroup.Use(SetCurrentCategory)
		catGroup.GET("/", CategoriesIndex).Name("cat")
		catGroup.GET("/createTopic",TopicCreateGet).Name("topicCreate")
		catGroup.POST("/createTopic",TopicCreatePost)

		topicGroup := catGroup.Group("/{tid}")

		topicGroup.GET("/",TopicGet).Name("topicGet") //

		//topicGroup.GET("/create",TopicCreateGet)
		//catGroup.GET("/create", CategoriesCreateGet)
		//catGroup.POST("/create", CategoriesCreatePost)
		//catGroup.GET("/detail", CategoriesDetail)

		admin := app.Group("/admin")
		admin.Use(SiteStruct)
		admin.GET("/f", manageForum)
		admin.GET("newforum",createForum)
		//admin.GET("newforum/post", createForumPost)
		admin.POST("newforum/post", createForumPost)

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "es-es"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

func NotFound(c buffalo.Context) error {
	return c.Render(404, r.HTML("meta/404.plush.html"))
}