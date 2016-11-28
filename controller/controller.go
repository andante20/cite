package controller

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	gologging "github.com/op/go-logging"
	"github.com/kakao/cite/models"
)

var (
	err    error
	logger = gologging.MustGetLogger("stdout")

	es             = models.NewElastic()
	k8s            = models.NewKubernetes()
	util           = models.NewUtil()
	buildbotClient = models.NewBuildBot()
	docker         = models.NewDocker()
	commonGitHub   = models.NewCommonGitHub()
	GMT, _         = time.LoadLocation("GMT")
	watchcenter    = models.NewWatchCenter()

	sessionStore = sessions.NewCookieStore([]byte("1VMo28DykUsIM1L8"))
)

func init() {
	sessionStore.Options = &sessions.Options{
		MaxAge: 3600,
		Path:   "/",
	}

	gob.Register(mapset.NewSet())
}

func getSession(c echo.Context) *sessions.Session {
	session, err := sessionStore.Get(c.Request().(*standard.Request).Request, models.Conf.Cite.Version)	
	if err != nil {
		logger.Error("failed to get session")
	}
	return session
}

func saveSession(session *sessions.Session, c echo.Context) error {
	err = session.Save(
		c.Request().(*standard.Request).Request,
		c.Response().(*standard.Response).ResponseWriter)
	if err != nil {
		logger.Errorf("failed to save session: %v", err)
		return err
	}
	return nil
}

func destroySession(c echo.Context) error {
	session := getSession(c)
	session.Options.MaxAge = -1
	return saveSession(session, c)
}

func AuthAPI(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := getSession(c)
		_, ok := session.Values["token"]
		if !ok {
			return echo.ErrUnauthorized
		}
		return next(c)
	}
}

func AuthWeb(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := getSession(c)
		_, ok := session.Values["token"]
		if !ok {
			session.Values["redirectPath"] = c.Request().URI()
			saveSession(session, c)
			return c.Redirect(http.StatusFound, "/login")
		}
		return next(c)
	}
}
