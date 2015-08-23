package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/render"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Running struct {
	User         string
	Date_str     string
	Distance     float64
	Elapsed_time int64
	Activities   int64
	Time_str     time.Duration
	Day          string
	Pace         time.Duration
}

type Runs struct {
	Id   bson.ObjectId `bson:"_id"`
	Date string        `bson:"date"`
	Run  string        `bson:"run"`
}

type Config struct {
	Host     string
	Username string
	Password string
	Database string
}

type Data struct {
	Rundata
}

type Rundata struct {
	Thisyear []string
	Lastyear []string
}

func handlerICon(w http.ResponseWriter, r *http.Request) {}

//Get totals
func getSum(host, user, pw, dbase, date string) (runs []Running) {
	db, err := sql.Open("mysql", user+":"+pw+"@tcp("+host+":3306)/"+dbase)
	defer db.Close()

	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	_, err = dbmap.Select(&runs, "select user, count(activity_id) as activities, sum(distance) as distance, "+
		" sum(elapsed_time) as elapsed_time from running where substring(date_str,1,7) = ? group by user order by 2 desc", date)

	if err != nil {
		fmt.Println("Not found", err)
	}

	for i := 0; i < len(runs); i++ {

		//Get total duration
		runs[i].Time_str = time.Duration(runs[i].Elapsed_time) * time.Second
		sekunder_float := float64(runs[i].Elapsed_time)
		//Get avg pace
		runs[i].Pace = time.Duration(sekunder_float/runs[i].Distance) * time.Second
	}
	return
}

//Get distance of each run
func getEach(host, user, pw, dbase, date string) (m map[string][]float64) {

	db, err := sql.Open("mysql", user+":"+pw+"@tcp("+host+":3306)/"+dbase)
	defer db.Close()

	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	var rns []Running
	_, err = dbmap.Select(&rns, "select d.name as 'user', d.day as 'day', ifnull(r.distance, 0) as 'distance'"+
		" from days d left join running r on substr(r.date_str,9,11) = d.day and r.user = d.name"+
		" and substring(r.date_str,1,7) = ? order by d.name, d.day", date)

	if err != nil {
		fmt.Println("Not found", err)
	}

	var name string
	m = make(map[string][]float64)

	for i := 0; i < len(rns); i++ {

		name = rns[i].User
		m[name] = append(m[name], rns[i].Distance)
		if err != nil {
			log.Println(err)
		}
	}

	return
}

//Get stats
func getYear(year string) []string {
	session, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		panic(err)
	}
	defer session.Close()

	c := session.DB("time").C("runs")

	var result []Runs
	err = c.Find(bson.M{"date": &bson.RegEx{Pattern: year, Options: "i"}}).Sort("date").All(&result)
	if err != nil {
		log.Fatal(err)
	}

	var acc_values []float64
	var string_values []string
	var dist_float float64
	for i := 0; i < len(result); i++ {

		//Convert to float
		if len(result[i].Run) > 0 {
			dist_float, err = strconv.ParseFloat(result[i].Run, 64)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			dist_float = 0
		}

		if i > 0 {
			prev_value := acc_values[i-1]
			current_value := float64(prev_value + dist_float)

			current_value_str := strconv.FormatFloat(current_value, 'f', 2, 64)
			acc_values = append(acc_values, current_value)
			string_values = append(string_values, current_value_str)

		} else {
			acc_values = append(acc_values, dist_float)
			dist_float_str := strconv.FormatFloat(dist_float, 'f', 2, 64)
			string_values = append(string_values, dist_float_str)
		}
	}
	return string_values
}

func main() {

	m := martini.Classic()
	m.Use(render.Renderer())

	http.HandleFunc("/favicon.ico", handlerICon)

	//Read config file
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	var config Config
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}

	m.Get("/", func(r render.Render) {

		month := time.Now().Month()
		date := time.Now().Format("2006-01")
		prev_month := time.Now().AddDate(0, -1, 0).Format("2006-01")
		current_month := time.Now().Format("2006-01")

		s := getSum(config.Host, config.Username, config.Password, config.Database, date)

		r.HTML(200, "templ", map[string]interface{}{"sum": s, "month": month, "prev_month": prev_month, "current_month": current_month, "date": date})
	})

	m.Get("/getruns/:date", func(params martini.Params, r render.Render) {

		match, _ := regexp.MatchString("[0-9]{4}-[0-9]{2}", params["date"])
		if match == true {
			date := params["date"]

			e := getEach(config.Host, config.Username, config.Password, config.Database, date)

			r.JSON(200, e)
		}

	})

	m.Get("/m/:date", func(params martini.Params, r render.Render) {

		match, _ := regexp.MatchString("[0-9]{4}-[0-9]{2}", params["date"])
		if match == true {
			layout := "2006-01"

			date_time, err := time.Parse(layout, params["date"])
			if err != nil {
				fmt.Println(err)
			}

			month := date_time.Month()
			date := params["date"]
			prev_month := date_time.AddDate(0, -1, 0).Format("2006-01")
			current_month := time.Now().Format("2006-01")

			s := getSum(config.Host, config.Username, config.Password, config.Database, date)

			r.HTML(200, "templ", map[string]interface{}{"sum": s, "month": month,
				"prev_month": prev_month, "current_month": current_month, "date": params["date"]})
		} else {
			r.Redirect("/")
		}

	})

	m.Get("/stats", func(r render.Render) {

		r.HTML(200, "stats", nil)
	})

	m.Get("/getstats", func(r render.Render) {

		this_year := getYear("2015")
		last_year := getYear("2014")

		data := Data{Rundata: Rundata{Thisyear: this_year, Lastyear: last_year}}
		r.JSON(200, data)
	})

	log.Fatal(http.ListenAndServe(":3010", m))

}
