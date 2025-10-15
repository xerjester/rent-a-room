package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Room struct {
	ID         int
	RoomNumber string
	Size       string
	Price      float64
	Image      string
	Status     string
}

type Booking struct {
	ID           int
	RoomNumber   string
	CustomerName string
	Phone        string
	StartDate    string
	EndDate      string
}

type BookingDisplay struct {
	RoomNumber   string
	CustomerName string
	Phone        string
	StartDate    string
	EndDate      string
}

func connectDB() (*sql.DB, error) {
	return sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/rental_system")
}

func ensureUploadDir() {
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
}

func main() {
	r := gin.Default()
	ensureUploadDir()
	r.Static("/uploads", "./uploads")

	r.SetFuncMap(template.FuncMap{
		"add1": func(i int) int { return i + 1 },
		"seq": func(start, end int) []int {
			s := make([]int, end-start+1)
			for i := range s {
				s[i] = start + i
			}
			return s
		},
	})
	r.LoadHTMLGlob("templates/*.html")

	// หน้าแรกแสดงห้องว่าง
	r.GET("/", func(c *gin.Context) {
		db, _ := connectDB()
		defer db.Close()

		rows, err := db.Query("SELECT id, room_number, size, price, image, status FROM rooms WHERE status='available'")
		if err != nil {
			c.String(http.StatusInternalServerError, "ดึงข้อมูลห้องล้มเหลว: "+err.Error())
			return
		}
		defer rows.Close()

		var rooms []Room
		for rows.Next() {
			var rm Room
			rows.Scan(&rm.ID, &rm.RoomNumber, &rm.Size, &rm.Price, &rm.Image, &rm.Status)
			rooms = append(rooms, rm)
		}

		c.HTML(http.StatusOK, "index.html", gin.H{"Rooms": rooms})
	})

	// Route book
	r.GET("/book/:id", func(c *gin.Context) {
		id := c.Param("id")
		db, _ := connectDB()
		defer db.Close()

		var room Room
		err := db.QueryRow("SELECT id, room_number, size, price, image FROM rooms WHERE id=?", id).
			Scan(&room.ID, &room.RoomNumber, &room.Size, &room.Price, &room.Image)
		if err != nil {
			c.String(http.StatusNotFound, "ไม่พบห้องที่เลือก")
			return
		}

		c.HTML(http.StatusOK, "book.html", gin.H{"room": room})
	})

	r.POST("/book", func(c *gin.Context) {
		roomIDStr := c.PostForm("room_id")
		name := c.PostForm("name")
		phone := c.PostForm("phone")
		startDateStr := c.PostForm("start_date")
		monthsStr := c.PostForm("months")

		roomID, _ := strconv.Atoi(roomIDStr)
		months, _ := strconv.Atoi(monthsStr)

		// แปลง start_date เป็น time.Time
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.String(http.StatusBadRequest, "รูปแบบวันที่เริ่มต้นไม่ถูกต้อง")
			return
		}

		// คำนวณ end_date
		endDate := startDate.AddDate(0, months, 0)

		db, _ := connectDB()
		defer db.Close()

		// ตรวจสอบสถานะห้องก่อน
		var status string
		err = db.QueryRow("SELECT status FROM rooms WHERE id=?", roomID).Scan(&status)
		if err != nil {
			c.String(http.StatusInternalServerError, "ไม่พบห้อง")
			return
		}
		if status != "available" {
			c.String(http.StatusBadRequest, "ห้องนี้ถูกจองไปแล้ว")
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.String(http.StatusInternalServerError, "เกิดข้อผิดพลาด")
			return
		}

		// บันทึกการจอง
		_, err = tx.Exec(
			"INSERT INTO bookings (room_id, customer_name, phone, start_date, end_date, months) VALUES (?, ?, ?, ?, ?, ?)",
			roomID, name, phone, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), months)
		if err != nil {
			tx.Rollback()
			c.String(http.StatusInternalServerError, "บันทึกการจองไม่สำเร็จ: "+err.Error())
			return
		}

		// อัปเดตสถานะห้องเป็น booked
		_, err = tx.Exec("UPDATE rooms SET status='booked' WHERE id=?", roomID)
		if err != nil {
			tx.Rollback()
			c.String(http.StatusInternalServerError, "อัปเดตสถานะห้องไม่สำเร็จ: "+err.Error())
			return
		}

		tx.Commit()

		c.HTML(http.StatusOK, "success.html", gin.H{
			"message": "จองห้องพักสำเร็จแล้ว!",
			"name":    name,
			"start":   startDate.Format("2006-01-02"),
			"end":     endDate.Format("2006-01-02"),
			"months":  months,
		})
	})

	// LOGIN
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		db, _ := connectDB()
		defer db.Close()

		var dbPass string
		err := db.QueryRow("SELECT password FROM admins WHERE username=?", username).Scan(&dbPass)
		if err != nil || dbPass != password {
			c.String(http.StatusUnauthorized, "Username หรือ Password ไม่ถูกต้อง")
			return
		}

		c.SetCookie("admin", username, 3600, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/dashboard")
	})

	// DASHBOARD
	r.GET("/dashboard", func(c *gin.Context) {
		admin, _ := c.Cookie("admin")
		if admin == "" {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		db, _ := connectDB()
		defer db.Close()

		// ดึงข้อมูลห้อง
		rooms := []Room{}
		rows, _ := db.Query("SELECT id, room_number, size, price, image FROM rooms")
		defer rows.Close()
		for rows.Next() {
			var rm Room
			rows.Scan(&rm.ID, &rm.RoomNumber, &rm.Size, &rm.Price, &rm.Image)
			rooms = append(rooms, rm)
		}

		// ดึงข้อมูลการจอง
		bookings := []Booking{}
		bRows, _ := db.Query(`
        SELECT b.id, r.room_number, b.customer_name, b.phone, b.start_date, b.end_date
        FROM bookings b
        JOIN rooms r ON b.room_id = r.id
        ORDER BY b.start_date ASC
    `)
		defer bRows.Close()
		for bRows.Next() {
			var b Booking
			bRows.Scan(&b.ID, &b.RoomNumber, &b.CustomerName, &b.Phone, &b.StartDate, &b.EndDate)
			bookings = append(bookings, b)
		}

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"Admin":    admin,
			"Rooms":    rooms,
			"Bookings": bookings,
		})
	})

	// DELETE BOOKING
	r.GET("/delete_booking/:id", func(c *gin.Context) {
		id := c.Param("id")

		db, _ := connectDB()
		defer db.Close()

		// ดึง room_id ของ booking ก่อน
		var roomID int
		err := db.QueryRow("SELECT room_id FROM bookings WHERE id=?", id).Scan(&roomID)
		if err != nil {
			c.String(http.StatusNotFound, "ไม่พบการจอง")
			return
		}

		// ลบ booking
		_, err = db.Exec("DELETE FROM bookings WHERE id=?", id)
		if err != nil {
			c.String(http.StatusInternalServerError, "ลบการจองล้มเหลว: "+err.Error())
			return
		}

		// อัปเดตสถานะห้องเป็น available
		_, err = db.Exec("UPDATE rooms SET status='available' WHERE id=?", roomID)
		if err != nil {
			c.String(http.StatusInternalServerError, "อัปเดตสถานะห้องล้มเหลว: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/dashboard")
	})

	// ------------------- ADD ROOM -------------------
	r.GET("/add_room", func(c *gin.Context) {
		admin, _ := c.Cookie("admin")
		if admin == "" {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		c.HTML(http.StatusOK, "add_room.html", nil)
	})

	r.POST("/add_room", func(c *gin.Context) {
		roomNumber := c.PostForm("room_number")
		size := c.PostForm("size")
		price, _ := strconv.ParseFloat(c.PostForm("price"), 64)

		file, err := c.FormFile("image")
		if err != nil {
			c.String(http.StatusBadRequest, "ต้องเลือกไฟล์รูปภาพ")
			return
		}

		filename := strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)
		c.SaveUploadedFile(file, savePath)

		db, _ := connectDB()
		defer db.Close()

		// ใช้ Prepared Statement ป้องกัน SQL Injection
		stmt, err := db.Prepare("INSERT INTO rooms (room_number, size, price, image, status) VALUES (?, ?, ?, ?, 'available')")
		if err != nil {
			c.String(http.StatusInternalServerError, "เกิดข้อผิดพลาด: "+err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(roomNumber, size, price, filename)
		if err != nil {
			c.String(http.StatusInternalServerError, "เพิ่มห้องล้มเหลว: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/dashboard")
	})

	// ------------------- EDIT ROOM -------------------
	r.GET("/edit_room/:id", func(c *gin.Context) {
		admin, _ := c.Cookie("admin")
		if admin == "" {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		id := c.Param("id")
		db, _ := connectDB()
		defer db.Close()

		var room Room
		err := db.QueryRow("SELECT id, room_number, size, price, image, status FROM rooms WHERE id=?", id).
			Scan(&room.ID, &room.RoomNumber, &room.Size, &room.Price, &room.Image, &room.Status)
		if err != nil {
			c.String(http.StatusNotFound, "ไม่พบห้องนี้")
			return
		}

		c.HTML(http.StatusOK, "edit_room.html", gin.H{"Room": room})
	})

	r.POST("/edit_room/:id", func(c *gin.Context) {
		id := c.Param("id")
		roomNumber := c.PostForm("room_number")
		size := c.PostForm("size")
		price, _ := strconv.ParseFloat(c.PostForm("price"), 64)

		db, _ := connectDB()
		defer db.Close()

		var filename string
		file, err := c.FormFile("image")
		if err == nil {
			filename = strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(file.Filename)
			savePath := filepath.Join("uploads", filename)
			c.SaveUploadedFile(file, savePath)
		} else {
			db.QueryRow("SELECT image FROM rooms WHERE id=?", id).Scan(&filename)
		}

		// ใช้ Prepared Statement ป้องกัน SQL Injection
		stmt, err := db.Prepare("UPDATE rooms SET room_number=?, size=?, price=?, image=? WHERE id=?")
		if err != nil {
			c.String(http.StatusInternalServerError, "เกิดข้อผิดพลาด: "+err.Error())
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(roomNumber, size, price, filename, id)
		if err != nil {
			c.String(http.StatusInternalServerError, "แก้ไขห้องล้มเหลว: "+err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/dashboard")
	})

	// ------------------- DELETE ROOM -------------------
	r.GET("/delete_room/:id", func(c *gin.Context) {
		id := c.Param("id")
		db, _ := connectDB()
		defer db.Close()

		// ลบ booking ที่เกี่ยวข้อง
		db.Exec("DELETE FROM bookings WHERE room_id=?", id)

		var oldImage string
		db.QueryRow("SELECT image FROM rooms WHERE id=?", id).Scan(&oldImage)
		if oldImage != "" {
			os.Remove(filepath.Join("uploads", oldImage))
		}

		db.Exec("DELETE FROM rooms WHERE id=?", id)
		c.Redirect(http.StatusFound, "/dashboard")
	})

	r.Run(":8080")
}
