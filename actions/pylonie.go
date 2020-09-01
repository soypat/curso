package actions

import (
	"crypto"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/soypat/curso/models"
	"go.etcd.io/bbolt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const nullUUID = "00000000-0000-0000-0000-000000000000"

var DB *bbolt.DB

func init() {
	var err error
	DB, err = bbolt.Open("tmp/codes.db", 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		defer panic(err)
		os.Exit(1)
	}
}


// recieve POST request to submit and evaluate code
func InterpretPost(c buffalo.Context) error {
	p := pythonHandler{}
	u := c.Value("current_user")
	if u  == nil {
		return c.Render(403,r.HTML("index.plush.html"))
	}
	user := u.(*models.User)
	p.UserName = user.Name
	p.userID = encode([]rune(user.ID.String()), b64safe)
	if err := c.Bind(&p.code); err != nil {
		_ = p.codeResult(c,"","An unexpected error occurred. You were logged out")
		return AuthDestroy(c)
	}
	if p.code.Evaluation.String() !=  nullUUID {
		if i,err:=strconv.ParseInt(p.code.Input,10,64); p.code.Input == "" || err!=nil || i<1000 {
			return p.codeResult(c,"",T.Translate(c,"curso-python-input-field-error"))
		}
		c.Logger().Info("starting evaluation!")
		tx := c.Value("tx").(*pop.Connection)
		q := tx.Where("id = ?", p.code.Evaluation)
		exists, err := q.Exists("evaluations")
		if err != nil  {
			return p.codeResult(c,"",T.Translate(c,"app-status-internal-error"))
		}
		if !exists {
			return p.codeResult(c,"",T.Translate(c,"curso-python-evaluation-not-found"))
		}
		eval := &models.Evaluation{}
		if err = q.First(eval); err != nil {
			return  p.codeResult(c,"",T.Translate(c,"curso-python-evaluation-not-found"))
		}
		peval := pythonHandler{}
		peval.userID = p.userID
		peval.Source = eval.Solution
		peval.Input = p.Input //eval.Inputs.String
		err = peval.runPy()
		if err != nil {
			return  p.codeResult(c,peval.Output,"Evaluation errored! "+err.Error()) // TODO this is the debug line
			//return  p.codeResult(c,"","Evaluation errored! "+err.Error()) // TODO this is the production line
		}
		defer p.Put(DB, c)
		p.Input = eval.Inputs.String
		err = p.runPy()
		if err != nil {
			return p.codeResult(c, "", err.Error())
		}
		if p.Output == peval.Output {
			return p.codeResult(c, T.Translate(c,"curso-python-evaluation-success"))
		} else {
			return p.codeResult(c, "", T.Translate(c,"curso-python-evaluation-fail"))
		}
	}
	defer p.Put(DB, c)
	err := p.runPy()
	if err != nil {
		return p.codeResult(c,p.result.Output, err.Error())
	}
	return p.codeResult(c)
}

// adds code result to context response.
// First and second string inputs will replace
// stdout and stderr code output, respectively
// so be careful not to delete important output/error
func (p *pythonHandler)codeResult(c buffalo.Context, output ...string) error {
	if len(output)==1 {
		p.result.Output = output[0]
	}
	if len(output)==2 {
		p.result.Output = output[0]
		p.result.Error = output[1]
	}
	jsonResponse, _ := json.Marshal(p.result)
	c.Response().Write(jsonResponse)
	return nil
}

// configuration values
const (
	pyCommand          = "python3"
	dbUploadBucketName = "uploads"
	pyTimeout_ms       = 2500
	// DB:
	pyMaxSourceLength  = 5000 // DB storage trim length
	pyMaxOutputLength  = 2000 // in characters
)

type pyExitStatus int

const (
	pyOK pyExitStatus = iota
	pyTimeout
	pyError
)

type code struct {
	Source string `json:"code" form:"code"`
	Input string `json:"input" form:"input"`
	Evaluation uuid.UUID `json:"evalid" form:"evalid"`
}

type result struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

type pythonHandler struct {
	result
	code
	Time     string `json:"time"`
	UserName string `json:"user"`
	userID   string
	filename string
}

var reForbid = map[*regexp.Regexp]string{
	regexp.MustCompile(`exec|open|write|eval|Write|globals|locals|breakpoint|getattr|memoryview|vars|super`): "forbidden function key '%s'",
	//regexp.MustCompile(`input\s*\(`):                           "no %s) to parse!",
	regexp.MustCompile("tofile|savetxt|fromfile|fromtxt|load"): "forbidden numpy function key '%s'",
	regexp.MustCompile(`__\w+__`):                              "forbidden dunder function key '%s'",
}

var reImport = regexp.MustCompile(`^from[\s]+[\w]+|import[\s]+[\w]+`)

var allowedImports = map[string]bool{
	"math":       true,
	"numpy":      true,
	"itertools":  true,
	"processing": false,
	"os":         false,
}

func (p *pythonHandler) runPy() (err error) {
	err = p.code.sanitizePy()
	output := make([]byte, 0)
	if err != nil {
		return
	}
	os.Mkdir(fmt.Sprintf("tmp/%s", p.userID), os.ModeTemporary)
	if err != nil && err != os.ErrExist {
		return
	}
	filename := fmt.Sprintf("tmp/%s/f.py", p.userID)

	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write([]byte(p.code.Source))
	cmd := exec.Command(pyCommand, filename)
	stdin, _ := cmd.StdinPipe()
	go func() {
		stdin.Write([]byte(p.Input+"\n"))
	}()
	status := make(chan pyExitStatus, 1)
	go func() {
		time.Sleep(pyTimeout_ms * time.Millisecond)
		status <- pyTimeout
	}()
	go func() {
		output, err = cmd.CombinedOutput()
		if err != nil {
			status <- pyError
		} else {
			status <- pyOK
		}
	}()
	for {
		select {
		case s := <-status:
			switch s {
			case pyTimeout:
				cmd.Process.Kill()
				return fmt.Errorf("process timed out (%dms)", pyTimeout_ms)
			case pyError, pyOK:
				p.Output = strings.ReplaceAll(string(output), "\""+filename+"\",", "")
				return
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}
}



func (c *code) sanitizePy() error {
	if len(c.Source) > 600 {
		return fmt.Errorf("code snippet too long!")
	}
	semicolonSplit := strings.Split(c.Source, ";")
	newLineSplit := strings.Split(c.Source, "\n")
	for _, v := range append(semicolonSplit, newLineSplit...) {
		for re, errF := range reForbid {
			str := re.FindString(strings.TrimSpace(v))
			if str != "" {
				return fmt.Errorf(errF, str)
			}
		}
		str := reImport.FindString(strings.TrimSpace(v))
		if str != "" {
			words := strings.Split(str, " ")
			if len(words) < 2 {
				return fmt.Errorf("unexpected import formatting: %s", str)
			}
			allowed, present := allowedImports[strings.TrimSpace(words[1])]
			if !present {
				return fmt.Errorf("import '%s' not in safelist:\n%s", strings.TrimSpace(words[1]), printSafeList())
			}
			if !allowed {
				return fmt.Errorf("forbidden import '%s'", strings.TrimSpace(words[1]))
			}
		}
	}
	return nil
}
func printSafeList() (s string) {
	counter := 0
	for k, v := range allowedImports {
		if v {
			counter++
			if counter > 1 {
				s += ",  "
			}
			s += k
		}
	}
	return
}

// Saves Python code and user to database
func (p *pythonHandler) Put(db *bbolt.DB, c buffalo.Context) {
	err := db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(dbUploadBucketName))
		if err != nil {
			return err
		}
		p.Time = time.Now().String()
		var pc pythonHandler
		pc = *p // because we don't want to store 5000000 length outputs
		if len(pc.Output) > pyMaxOutputLength {
			pc.Output = pc.Output[:pyMaxOutputLength]
		}
		if len(pc.Source) > pyMaxSourceLength {
			pc.Source = pc.Source[:pyMaxSourceLength]
		}
		buff, err := json.Marshal(pc)
		if err != nil {
			return err
		}
		h := crypto.MD5.New()
		h.Write([]byte(pc.UserName + pc.code.Source))
		sum := h.Sum(nil)
		if b.Get(sum) == nil {
			c.Logger().Infof("Code submitted user: %s", pc.UserName)
			return b.Put(h.Sum(nil), buff)
		}
		c.Logger().Infof("Repeated code input submitted user: %s", pc.UserName)
		return nil
	})

	if err != nil {
		c.Logger().Errorf("could not save python code to database for user '%s'\n", p.UserName)
	}
}

// if accessed by registered user with 'admin' role then db containing code is downloaded
func pyDBBackup(c buffalo.Context) error {
	w := c.Response()
	err := DB.View(func(tx *bbolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="uploads.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return nil
}
