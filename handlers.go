package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func jsonResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{Code: code, Message: message, Data: data})
}

func handleTanks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tanks, err := GetAllTanks()
		if err != nil {
			jsonResponse(w, 500, "获取鱼缸列表失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "success", tanks)
	case "POST":
		var t Tank
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if t.TankNo == "" || t.Type == "" || t.Volume <= 0 || t.Location == "" {
			jsonResponse(w, 400, "参数不完整或不合法", nil)
			return
		}
		if t.Status == "" {
			t.Status = "正常"
		}
		if err := AddTank(t); err != nil {
			jsonResponse(w, 500, "添加鱼缸失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "添加成功", t)
	case "PUT":
		var t Tank
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if err := UpdateTank(t); err != nil {
			jsonResponse(w, 500, "更新鱼缸失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "更新成功", t)
	case "DELETE":
		tankNo := r.URL.Query().Get("tank_no")
		if tankNo == "" {
			jsonResponse(w, 400, "缺少缸号参数", nil)
			return
		}
		if err := DeleteTank(tankNo); err != nil {
			jsonResponse(w, 500, "删除鱼缸失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "删除成功", nil)
	default:
		jsonResponse(w, 405, "方法不允许", nil)
	}
}

func handleCreatures(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		creatures, err := GetAllCreatures()
		if err != nil {
			jsonResponse(w, 500, "获取生物列表失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "success", creatures)
	case "POST":
		var c Creature
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if c.Name == "" || c.Type == "" || c.TankNo == "" || c.EnterDate == "" || c.Quantity <= 0 {
			jsonResponse(w, 400, "参数不完整或不合法", nil)
			return
		}
		if err := AddCreature(c); err != nil {
			jsonResponse(w, 500, "添加生物失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "添加成功", c)
	case "PUT":
		var c Creature
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if err := UpdateCreature(c); err != nil {
			jsonResponse(w, 500, "更新生物失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "更新成功", c)
	case "DELETE":
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			jsonResponse(w, 400, "ID参数错误", nil)
			return
		}
		if err := DeleteCreature(id); err != nil {
			jsonResponse(w, 500, "删除生物失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "删除成功", nil)
	default:
		jsonResponse(w, 405, "方法不允许", nil)
	}
}

func handleWaterQuality(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tankNo := r.URL.Query().Get("tank_no")
		if tankNo == "" {
			jsonResponse(w, 400, "缺少缸号参数", nil)
			return
		}
		records, err := GetRecentWaterQuality(tankNo, 7)
		if err != nil {
			jsonResponse(w, 500, "获取水质记录失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "success", records)
	case "POST":
		var wq WaterQuality
		if err := json.NewDecoder(r.Body).Decode(&wq); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if wq.TankNo == "" || wq.TestDate == "" || wq.Recorder == "" {
			jsonResponse(w, 400, "参数不完整", nil)
			return
		}
		if wq.PH < 6.0 || wq.PH > 8.5 {
			jsonResponse(w, 400, "pH值必须在6.0-8.5之间", nil)
			return
		}
		if wq.Nitrite < 0 {
			jsonResponse(w, 400, "亚硝酸盐不能为负数", nil)
			return
		}
		if wq.Temperature < 18 || wq.Temperature > 32 {
			jsonResponse(w, 400, "温度必须在18-32℃之间", nil)
			return
		}
		tank, err := GetTankByNo(wq.TankNo)
		if err != nil {
			jsonResponse(w, 500, "查询鱼缸失败: "+err.Error(), nil)
			return
		}
		if tank == nil {
			jsonResponse(w, 400, "鱼缸不存在", nil)
			return
		}
		if tank.Status != "正常" {
			jsonResponse(w, 400, "只有正常状态的鱼缸才能添加水质记录", nil)
			return
		}
		if err := AddWaterQuality(wq); err != nil {
			jsonResponse(w, 500, "添加水质记录失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "添加成功", wq)
	default:
		jsonResponse(w, 405, "方法不允许", nil)
	}
}

func handleWaterChanges(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		records, err := GetAllWaterChanges()
		if err != nil {
			jsonResponse(w, 500, "获取换水记录失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "success", records)
	case "POST":
		var wc WaterChange
		if err := json.NewDecoder(r.Body).Decode(&wc); err != nil {
			jsonResponse(w, 400, "请求参数错误: "+err.Error(), nil)
			return
		}
		if wc.TankNo == "" || wc.Date == "" || wc.Volume <= 0 || wc.Operator == "" {
			jsonResponse(w, 400, "参数不完整或不合法", nil)
			return
		}
		tank, err := GetTankByNo(wc.TankNo)
		if err != nil {
			jsonResponse(w, 500, "查询鱼缸失败: "+err.Error(), nil)
			return
		}
		if tank == nil {
			jsonResponse(w, 400, "鱼缸不存在", nil)
			return
		}
		if err := AddWaterChange(wc); err != nil {
			jsonResponse(w, 500, "添加换水记录失败: "+err.Error(), nil)
			return
		}
		jsonResponse(w, 200, "添加成功", wc)
	default:
		jsonResponse(w, 405, "方法不允许", nil)
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonResponse(w, 405, "方法不允许", nil)
		return
	}
	stats, err := GetMonthlyWaterChangeStats()
	if err != nil {
		jsonResponse(w, 500, "获取统计数据失败: "+err.Error(), nil)
		return
	}
	jsonResponse(w, 200, "success", stats)
}

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/tanks", handleTanks)
	mux.HandleFunc("/api/creatures", handleCreatures)
	mux.HandleFunc("/api/water-quality", handleWaterQuality)
	mux.HandleFunc("/api/water-changes", handleWaterChanges)
	mux.HandleFunc("/api/stats/monthly-water-changes", handleStats)

	fs := http.FileServer(http.Dir("static"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			jsonResponse(w, 404, "接口不存在", nil)
			return
		}
		fs.ServeHTTP(w, r)
	})
}
