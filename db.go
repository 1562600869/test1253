package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Tank struct {
	ID       int    `json:"id"`
	TankNo   string `json:"tank_no"`
	Type     string `json:"type"`
	Volume   int    `json:"volume"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

type Creature struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	TankNo     string `json:"tank_no"`
	EnterDate  string `json:"enter_date"`
	Quantity   int    `json:"quantity"`
}

type WaterQuality struct {
	ID            int     `json:"id"`
	TankNo        string  `json:"tank_no"`
	TestDate      string  `json:"test_date"`
	PH            float64 `json:"ph"`
	Nitrite       float64 `json:"nitrite"`
	Temperature   float64 `json:"temperature"`
	Recorder      string  `json:"recorder"`
}

type WaterChange struct {
	ID         int    `json:"id"`
	TankNo     string `json:"tank_no"`
	Date       string `json:"date"`
	Volume     int    `json:"volume"`
	Operator   string `json:"operator"`
}

type TankTypeStats struct {
	TankType string `json:"tank_type"`
	Count    int    `json:"count"`
}

func InitDB(dataSource string) error {
	var err error
	db, err = sql.Open("sqlite3", dataSource)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	return createTables()
}

func createTables() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS tanks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tank_no TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		volume INTEGER NOT NULL,
		location TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT '正常'
	);
	CREATE TABLE IF NOT EXISTS creatures (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		tank_no TEXT NOT NULL,
		enter_date TEXT NOT NULL,
		quantity INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS water_quality (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tank_no TEXT NOT NULL,
		test_date TEXT NOT NULL,
		ph REAL NOT NULL,
		nitrite REAL NOT NULL,
		temperature REAL NOT NULL,
		recorder TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS water_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tank_no TEXT NOT NULL,
		date TEXT NOT NULL,
		volume INTEGER NOT NULL,
		operator TEXT NOT NULL
	);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func GetAllTanks() ([]Tank, error) {
	rows, err := db.Query("SELECT id, tank_no, type, volume, location, status FROM tanks ORDER BY tank_no")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tanks []Tank
	for rows.Next() {
		var t Tank
		err := rows.Scan(&t.ID, &t.TankNo, &t.Type, &t.Volume, &t.Location, &t.Status)
		if err != nil {
			return nil, err
		}
		tanks = append(tanks, t)
	}
	return tanks, nil
}

func GetTankByNo(tankNo string) (*Tank, error) {
	var t Tank
	err := db.QueryRow("SELECT id, tank_no, type, volume, location, status FROM tanks WHERE tank_no = ?", tankNo).
		Scan(&t.ID, &t.TankNo, &t.Type, &t.Volume, &t.Location, &t.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func AddTank(t Tank) error {
	_, err := db.Exec("INSERT INTO tanks (tank_no, type, volume, location, status) VALUES (?, ?, ?, ?, ?)",
		t.TankNo, t.Type, t.Volume, t.Location, t.Status)
	return err
}

func UpdateTank(t Tank) error {
	_, err := db.Exec("UPDATE tanks SET type=?, volume=?, location=?, status=? WHERE tank_no=?",
		t.Type, t.Volume, t.Location, t.Status, t.TankNo)
	return err
}

func DeleteTank(tankNo string) error {
	_, err := db.Exec("DELETE FROM tanks WHERE tank_no = ?", tankNo)
	return err
}

func GetAllCreatures() ([]Creature, error) {
	rows, err := db.Query("SELECT id, name, type, tank_no, enter_date, quantity FROM creatures ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creatures []Creature
	for rows.Next() {
		var c Creature
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.TankNo, &c.EnterDate, &c.Quantity)
		if err != nil {
			return nil, err
		}
		creatures = append(creatures, c)
	}
	return creatures, nil
}

func AddCreature(c Creature) error {
	_, err := db.Exec("INSERT INTO creatures (name, type, tank_no, enter_date, quantity) VALUES (?, ?, ?, ?, ?)",
		c.Name, c.Type, c.TankNo, c.EnterDate, c.Quantity)
	return err
}

func UpdateCreature(c Creature) error {
	_, err := db.Exec("UPDATE creatures SET name=?, type=?, tank_no=?, enter_date=?, quantity=? WHERE id=?",
		c.Name, c.Type, c.TankNo, c.EnterDate, c.Quantity, c.ID)
	return err
}

func DeleteCreature(id int) error {
	_, err := db.Exec("DELETE FROM creatures WHERE id = ?", id)
	return err
}

func AddWaterQuality(wq WaterQuality) error {
	_, err := db.Exec("INSERT INTO water_quality (tank_no, test_date, ph, nitrite, temperature, recorder) VALUES (?, ?, ?, ?, ?, ?)",
		wq.TankNo, wq.TestDate, wq.PH, wq.Nitrite, wq.Temperature, wq.Recorder)
	return err
}

func GetRecentWaterQuality(tankNo string, limit int) ([]WaterQuality, error) {
	rows, err := db.Query("SELECT id, tank_no, test_date, ph, nitrite, temperature, recorder FROM water_quality WHERE tank_no = ? ORDER BY test_date DESC, id DESC LIMIT ?", tankNo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []WaterQuality
	for rows.Next() {
		var w WaterQuality
		err := rows.Scan(&w.ID, &w.TankNo, &w.TestDate, &w.PH, &w.Nitrite, &w.Temperature, &w.Recorder)
		if err != nil {
			return nil, err
		}
		records = append(records, w)
	}
	return records, nil
}

func AddWaterChange(wc WaterChange) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE tanks SET status = '换水中' WHERE tank_no = ?", wc.TankNo)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO water_changes (tank_no, date, volume, operator) VALUES (?, ?, ?, ?)",
		wc.TankNo, wc.Date, wc.Volume, wc.Operator)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE tanks SET status = '正常' WHERE tank_no = ?", wc.TankNo)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetMonthlyWaterChangeStats() ([]TankTypeStats, error) {
	now := time.Now()
	year, month, _ := now.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	lastDay := time.Date(year, month+1, 0, 23, 59, 59, 0, time.Local).Format("2006-01-02 15:04:05")

	rows, err := db.Query(`
		SELECT t.type as tank_type, COUNT(wc.id) as count
		FROM tanks t
		LEFT JOIN water_changes wc ON t.tank_no = wc.tank_no AND wc.date >= ? AND wc.date <= ?
		GROUP BY t.type
	`, firstDay, lastDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []TankTypeStats
	for rows.Next() {
		var s TankTypeStats
		err := rows.Scan(&s.TankType, &s.Count)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func GetAllWaterChanges() ([]WaterChange, error) {
	rows, err := db.Query("SELECT id, tank_no, date, volume, operator FROM water_changes ORDER BY date DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []WaterChange
	for rows.Next() {
		var w WaterChange
		err := rows.Scan(&w.ID, &w.TankNo, &w.Date, &w.Volume, &w.Operator)
		if err != nil {
			return nil, err
		}
		records = append(records, w)
	}
	return records, nil
}
