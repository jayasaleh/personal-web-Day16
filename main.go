package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Project struct {
	Id              int64
	ProjectName     string
	StartDate       time.Time
	EndDate         time.Time
	StartFormat     string
	EndFormat       string
	DurationProject string
	Description     string
	Technologies    []string
	Image           string
	Author          string
}

type Tech struct {
	Tnode       bool
	Tnext       bool
	Treach      bool
	Ttypescript bool
}
type User struct {
	IdUser   int
	NameUser string
	Email    string
	Password string
}

var dataProject = []Project{
	{
		ProjectName: "Dumbways.id Project",
		Description: "Project personal web",
	},
}

func main() {
	e := echo.New()
	connection.DatabaseConnect()
	// akses tampilan
	e.Static("/assets", "assets")
	e.Static("/upload", "upload")
	//session
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("session"))))
	// routing //
	e.GET("/", home)
	e.GET("/addProject", addProject)
	e.GET("/contact", contact)
	e.GET("/detailProject/:id", projectDetail)
	e.GET("/delete/:id", deleteProject)
	e.GET("/updateProject/:id", updateProject)
	e.GET("/register", register)
	e.GET("/login", login)
	e.GET("/logoutUser", logoutUser)
	//session

	e.POST("/addRegister", addRegister)
	e.POST("/loginUser", loginUser)
	e.POST("/add-project", middleware.UploadFile(addDataProject))
	e.POST("/updateProject/:id", middleware.UploadFile(updateDataProject))
	e.Logger.Fatal(e.Start("localhost:5000"))
}

func home(c echo.Context) error {
	sess, _ := session.Get("session", c)

	var tmpl, err = template.ParseFiles("index.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	var result []Project
	if sess.Values["isLogin"] == true {
		projectdata, _ := connection.Conn.Query(context.Background(), "SELECT tb_projects.id, name, star_date, end_date, technologies,description, image, tb_users.name_user AS author FROM tb_projects INNER JOIN tb_users ON tb_projects.author_id= tb_users.id_user WHERE tb_projects.author_id=$1 ORDER BY tb_projects.id", sess.Values["id"])
		for projectdata.Next() {
			var each = Project{}

			err := projectdata.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Technologies, &each.Description, &each.Image, &each.Author)
			if err != nil {
				fmt.Println(err.Error())
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}

			each.DurationProject = CountDuration(each.StartDate, each.EndDate)
			result = append(result, each)
		}
	} else {
		projectdata, _ := connection.Conn.Query(context.Background(), "SELECT tb_projects.id, name, star_date, end_date, technologies,description, image, tb_users.name_user AS author FROM tb_projects INNER JOIN tb_users ON tb_projects.author_id= tb_users.id_user")
		for projectdata.Next() {
			var each = Project{}

			err := projectdata.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Technologies, &each.Description, &each.Image, &each.Author)
			if err != nil {
				fmt.Println(err.Error())
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
			}
			each.DurationProject = CountDuration(each.StartDate, each.EndDate)
			result = append(result, each)
		}
	}

	//map(tipe data) => key and value
	datas := map[string]interface{}{
		"Project":      result,
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"FlashName":    sess.Values["name"],
		"FlashId":      sess.Values["id"],
		"FlashLogin":   sess.Values["isLogin"],
	}

	delete(sess.Values, "message")
	delete(sess.Values, "status")

	sess.Save(c.Request(), c.Response())
	return tmpl.Execute(c.Response(), datas)
}

func addProject(c echo.Context) error {
	sess, _ := session.Get("session", c)
	datas := map[string]interface{}{
		"FlashName":  sess.Values["name"],
		"FlashId":    sess.Values["id"],
		"FlashLogin": sess.Values["isLogin"],
	}
	var tmpl, err = template.ParseFiles("myProject.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message ": err.Error()})
	}

	return tmpl.Execute(c.Response(), datas)
}

func contact(c echo.Context) error {
	sess, _ := session.Get("session", c)
	datas := map[string]interface{}{
		"FlashName":  sess.Values["name"],
		"FlashId":    sess.Values["id"],
		"FlashLogin": sess.Values["isLogin"],
	}
	var tmpl, err = template.ParseFiles("contact.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	return tmpl.Execute(c.Response(), datas)
}

func projectDetail(c echo.Context) error {
	sess, _ := session.Get("session", c)
	id, _ := strconv.Atoi(c.Param("id"))

	var tmpl, err = template.ParseFiles("detailProject.html")

	var Detail = Project{}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_projects WHERE id= $1", id).Scan(&Detail.Id, &Detail.ProjectName, &Detail.StartDate, &Detail.EndDate, &Detail.Description, &Detail.Technologies, &Detail.Image, &Detail.Author)

	Detail.DurationProject = CountDuration(Detail.StartDate, Detail.EndDate)
	Detail.StartFormat = GetDateFormat(Detail.StartDate)
	Detail.EndFormat = GetDateFormat(Detail.EndDate)

	data := map[string]interface{}{
		"Project":    Detail,
		"FlashName":  sess.Values["name"],
		"FlashId":    sess.Values["id"],
		"FlashLogin": sess.Values["isLogin"],
	}

	return tmpl.Execute(c.Response(), data)
}

func register(c echo.Context) error {
	sess, _ := session.Get("session", c)
	datas := map[string]interface{}{
		"FlashName":  sess.Values["name"],
		"FlashId":    sess.Values["id"],
		"FlashLogin": sess.Values["isLogin"],
	}
	var tmpl, err = template.ParseFiles("register.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	return tmpl.Execute(c.Response(), datas)
}

func addRegister(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_users (name_user, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message": err.Error()})
	}
	if err != nil {
		redirectWithMessage(c, "Register failed, please try again", false, "/register")
	}
	return redirectWithMessage(c, "Register Success", true, "/login")
}

func login(c echo.Context) error {

	var tmpl, err = template.ParseFiles("login.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	sess, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
		"FlashName":    sess.Values["name"],
		"FlashId":      sess.Values["id"],
		"FlashLogin":   sess.Values["isLogin"],
	}

	delete(sess.Values, "message")
	delete(sess.Values, "status")

	sess.Save(c.Request(), c.Response())
	return tmpl.Execute(c.Response(), flash)
}

func loginUser(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	email := c.FormValue("email")
	password := c.FormValue("password")

	user := User{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_users WHERE email=$1", email).Scan(&user.IdUser, &user.NameUser, &user.Email, &user.Password)
	if err != nil {
		return redirectWithMessage(c, "Email Salah!", false, "/login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return redirectWithMessage(c, "Password Salah!", false, "/login")
	}

	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = 10800
	sess.Values["message"] = "Login Success"
	sess.Values["status"] = true
	sess.Values["name"] = user.NameUser
	sess.Values["id"] = user.IdUser
	sess.Values["isLogin"] = true

	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func logoutUser(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func addDataProject(c echo.Context) error {
	sess, _ := session.Get("session", c)
	author := sess.Values["id"]
	projectName := c.FormValue("projectName")
	desc := c.FormValue("desc")
	start := c.FormValue("start")
	end := c.FormValue("end")
	layout := "2006-01-02"
	start_date, _ := time.Parse(layout, start)
	end_date, _ := time.Parse(layout, end)
	node := c.FormValue("node")
	next := c.FormValue("next")
	reach := c.FormValue("reach")
	typeScript := c.FormValue("typeScript")
	image := c.Get("dataFile").(string)

	var techList = []string{}
	if node != "" {
		techList = append(techList, node)
	}
	if next != "" {
		techList = append(techList, next)
	}
	if reach != "" {
		techList = append(techList, reach)
	}
	if typeScript != "" {
		techList = append(techList, typeScript)
	}

	var addData = Project{
		ProjectName:  projectName,
		Description:  desc,
		StartDate:    start_date,
		EndDate:      end_date,
		Technologies: techList,
		Image:        image,
	}

	sqlQuery := "INSERT INTO tb_projects (name, star_date, end_date, description, technologies, image, author_id) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := connection.Conn.Exec(context.Background(), sqlQuery, addData.ProjectName, addData.StartDate, addData.EndDate, addData.Description, addData.Technologies, addData.Image, author)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func CountDuration(start time.Time, end time.Time) string {
	timeDifferent := float64(end.Sub(start).Seconds())
	years := int(math.Floor(timeDifferent / (12 * 30 * 24 * 60 * 60)))
	months := int(math.Floor(timeDifferent / (30 * 24 * 60 * 60)))
	weeks := int(math.Floor(timeDifferent / (7 * 24 * 60 * 60)))
	days := int(math.Floor(timeDifferent / (24 * 60 * 60)))

	if years > 0 {
		str := strconv.Itoa(years) + " Tahun"
		return str
	}
	if months > 0 {
		str := strconv.Itoa(months) + " Bulan"
		return str
	}
	if weeks > 0 {
		str := strconv.Itoa(weeks) + " Minggu"
		return str
	}
	if days > 0 {
		str := strconv.Itoa(days) + " Hari"
		return str
	}
	return "cannot get duration"
}

func deleteProject(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id = $1", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func updateProject(c echo.Context) error {
	var tmpl, err = template.ParseFiles("updateProject.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message:": err.Error()})
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var ViewEdit = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, star_date, end_date, description, technologies, image FROM tb_projects WHERE id= $1", id).Scan(&ViewEdit.Id, &ViewEdit.ProjectName, &ViewEdit.StartDate, &ViewEdit.EndDate, &ViewEdit.Description, &ViewEdit.Technologies, &ViewEdit.Image)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}
	ViewEdit.StartFormat = InputHtmlDateFormat(ViewEdit.StartDate)
	ViewEdit.EndFormat = InputHtmlDateFormat(ViewEdit.EndDate)

	var techList = Tech{
		Tnode:       false,
		Tnext:       false,
		Treach:      false,
		Ttypescript: false,
	}
	for _, v := range ViewEdit.Technologies {
		if v == "node" {
			techList.Tnode = true
		}
		if v == "next" {
			techList.Tnext = true
		}
		if v == "reach" {
			techList.Treach = true
		}
		if v == "typeScript" {
			techList.Ttypescript = true
		}
	}

	data := map[string]interface{}{
		"Project": ViewEdit,
		"Tech":    techList,
	}
	return tmpl.Execute(c.Response(), data)
}

func updateDataProject(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	projectName := c.FormValue("projectName")
	desc := c.FormValue("desc")
	start := c.FormValue("start")
	end := c.FormValue("end")
	layout := "2006-01-02"
	start_date, _ := time.Parse(layout, start)
	end_date, _ := time.Parse(layout, end)
	node := c.FormValue("node")
	next := c.FormValue("next")
	reach := c.FormValue("reach")
	typeScript := c.FormValue("typeScript")
	image := c.Get("dataFile").(string)
	var techList = []string{}
	if node != "" {
		techList = append(techList, node)
	}
	if next != "" {
		techList = append(techList, next)
	}
	if reach != "" {
		techList = append(techList, reach)
	}
	if typeScript != "" {
		techList = append(techList, typeScript)
	}

	var UpdateData = Project{
		ProjectName:  projectName,
		Description:  desc,
		StartDate:    start_date,
		EndDate:      end_date,
		Technologies: techList,
		Image:        image,
	}

	_, err := connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET name = $1, star_date = $2, end_date = $3, description = $4, technologies = $5, image =$6 WHERE id = $7 ", UpdateData.ProjectName, UpdateData.StartDate, UpdateData.EndDate, UpdateData.Description, UpdateData.Technologies, UpdateData.Image, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"Message ": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func GetDateFormat(t time.Time) string {
	day := strconv.Itoa(t.Day())
	month := t.Month()
	year := strconv.Itoa(t.Year())
	result := fmt.Sprintf("%s %s %s", day, month, year)
	return result
}

func InputHtmlDateFormat(t time.Time) string {
	day := strconv.Itoa(t.Day())
	month := strconv.Itoa(int(t.Month()))
	year := strconv.Itoa(t.Year())
	result := fmt.Sprintf("%s-%02s-%02s", year, month, day)
	return result
}

func redirectWithMessage(c echo.Context, message string, status bool, path string) error {
	sess, _ := session.Get("session", c)
	sess.Values["message"] = message
	sess.Values["status"] = status
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, path)
}
