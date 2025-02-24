package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	entry()
}

func utcToLocalPreservingTime(t time.Time) time.Time {
	return t
	// // Get the UTC components from the original time.
	// year, month, day := t.UTC().Date()
	// hour, min, sec := t.UTC().Clock()
	// nsec := t.UTC().Nanosecond()
	// // Create a new time with the same components, but in the local time zone.
	// return time.Date(year, month, day, hour, min, sec, nsec, time.Local)
}

type Shift struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start"`
	EndTime   time.Time `json:"end"`
	Name      string    `json:"name"`
}

func entry() {
	db, err := sql.Open("sqlite3", "./data.db?_journal_mode=WAL&_foreign_keys=ON&mode=shared")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Create the shifts table if it doesn't exist
	createTableStmt := `
	CREATE TABLE IF NOT EXISTS shifts (
		id TEXT PRIMARY KEY,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		name TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTableStmt); err != nil {
		fmt.Println(err)
		return
	}

	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get start of week
		weekDay := time.Now().Weekday()
		startDay := time.Now().Add(-time.Duration(weekDay) * 24 * time.Hour)
		startOfWeek := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.Local)
		endOfNextWeek := startOfWeek.AddDate(0, 0, 14)
		fmt.Println(startOfWeek, endOfNextWeek)

		stmt := `SELECT start_time, end_time, name, id FROM shifts WHERE start_time BETWEEN ? AND ?`

		rows, err := db.Query(stmt, startOfWeek, endOfNextWeek)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		shifts := make([]Shift, 0)
		for rows.Next() {
			var shift Shift
			if err := rows.Scan(&shift.StartTime, &shift.EndTime, &shift.Name, &shift.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shifts = append(shifts, shift)
		}

		totalHours := 0.0
		shiftData := make(map[string][]Shift)
		for _, shift := range shifts {
			shift.StartTime = shift.StartTime.Local()
			shift.EndTime = shift.EndTime.Local()
			totalHours += shift.EndTime.Sub(shift.StartTime).Hours()
			if shift.StartTime.Day() == shift.EndTime.Day() {
				shiftData[shift.StartTime.Format(time.DateOnly)] = append(shiftData[shift.StartTime.Format(time.DateOnly)], shift)
			} else {
				// split shift into two shifts
				startShift := Shift{
					StartTime: shift.StartTime,
					EndTime:   time.Date(shift.StartTime.Year(), shift.StartTime.Month(), shift.StartTime.Day(), 23, 59, 59, 999999999, shift.StartTime.Location()),
					Name:      shift.Name,
					ID:        shift.ID,
				}
				endShift := Shift{
					StartTime: time.Date(shift.EndTime.Year(), shift.EndTime.Month(), shift.EndTime.Day(), 0, 0, 0, 0, shift.EndTime.Location()),
					EndTime:   shift.EndTime,
					Name:      shift.Name,
					ID:        shift.ID,
				}
				shiftData[startShift.StartTime.Format(time.DateOnly)] = append(shiftData[startShift.StartTime.Format(time.DateOnly)], startShift)
				shiftData[endShift.StartTime.Format(time.DateOnly)] = append(shiftData[endShift.StartTime.Format(time.DateOnly)], endShift)
			}
		}

		b, _ := json.MarshalIndent(shiftData, "", "  ")

		fmt.Println(string(b))

		t := template.New("")
		t = t.Funcs(template.FuncMap{
			"dayOfWeek": func(date string) string {
				t, _ := time.Parse(time.DateOnly, date)
				return t.Weekday().String()
			},
			"timeOnly": func(t time.Time) string {
				return t.Local().Format("15:04")
			},
		})
		t, err = t.ParseGlob("web/*.tmpl")
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := t.ExecuteTemplate(w, "index.tmpl", map[string]interface{}{"ShiftData": shiftData, "TotalHours": totalHours}); err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		t := template.New("")
		t = t.Funcs(template.FuncMap{
			"dayOfWeek": func(date string) string {
				fmt.Println(date)
				t, _ := time.Parse(time.DateOnly, date)
				return t.Local().Weekday().String()
			},
			"timeOnly": func(t time.Time) string {
				return t.Local().Format("15:04")
			},
		})
		t, err := t.ParseGlob("web/*.tmpl")
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := t.ExecuteTemplate(w, "install.tmpl", nil); err != nil {
			fmt.Println(err)
		}

	})

	mux.HandleFunc("POST /api/shifts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("POST /api/shifts")

		var shifts []Shift
		if err := json.NewDecoder(r.Body).Decode(&shifts); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(shifts)
		for _, shift := range shifts {
			shift.EndTime = utcToLocalPreservingTime(shift.EndTime)
			shift.StartTime = utcToLocalPreservingTime(shift.StartTime)
			// insert or replace shift if id exists
			stmt := `INSERT INTO shifts (start_time, end_time, name, id) VALUES (?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET start_time = ?, end_time = ?, name = ?`

			if _, err := db.Exec(stmt, shift.StartTime, shift.EndTime, shift.Name, shift.ID, shift.StartTime, shift.EndTime, shift.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Write([]byte("Shift created successfully"))
	})

	mux.HandleFunc("GET /api/shifts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET /api/shifts")

		// get start of week
		weekDay := time.Now().Weekday()
		startDay := time.Now().Add(-time.Duration((int(weekDay)+6)%7) * 24 * time.Hour)
		startOfWeek := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.Local)
		endOfNextWeek := startOfWeek.AddDate(0, 0, 14)
		fmt.Println(startOfWeek, endOfNextWeek)

		stmt := `SELECT start_time, end_time, name, id FROM shifts WHERE start_time BETWEEN ? AND ?`

		rows, err := db.Query(stmt, startOfWeek, endOfNextWeek)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		shifts := make([]Shift, 0)
		for rows.Next() {
			var shift Shift
			if err := rows.Scan(&shift.StartTime, &shift.EndTime, &shift.Name, &shift.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shifts = append(shifts, shift)
		}

		if err := json.NewEncoder(w).Encode(shifts); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Server started at http://localhost:3000")
	go func() {
		if err := http.ListenAndServeTLS(":3001", "./server.crt", "./server.key", mux); err != nil {
			fmt.Println(err)
		}
	}()

	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println(err)
	}
}
