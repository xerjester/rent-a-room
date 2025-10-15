🏠 Room Rental System (ระบบเช่าห้องพัก)

ระบบเช่าห้องพักพัฒนาโดยใช้ภาษา Go (Golang) และ Gin Framework เชื่อมต่อฐานข้อมูล MySQL สำหรับจัดการข้อมูลห้องพัก การจอง และผู้ดูแลระบบ (Admin)

------------------------------------------------------------
🚀 ฟีเจอร์หลัก

👨‍💼 ฝั่งผู้ดูแลระบบ (Admin)
- เข้าสู่ระบบ (Login)
- เพิ่ม / แก้ไข / ลบ ห้องพัก
- อัปโหลดรูปภาพห้องพัก
- ดูสถานะห้อง (ว่าง / ถูกจองแล้ว)
- เปลี่ยนสถานะห้องอัตโนมัติเมื่อมีการจอง

🧍‍♂️ ฝั่งผู้เช่า (User)
- ดูห้องพักทั้งหมด
- กรอกข้อมูลเพื่อจองห้อง
- ระบบป้องกันการจองทับ (ไม่สามารถจองห้องที่ถูกจองอยู่แล้วได้)
- ระบบป้องกัน SQL Injection ด้วย Prepared Statement
- ตรวจสอบสถานะห้องหลังจากจองสำเร็จ

------------------------------------------------------------
⚙️ เทคโนโลยีที่ใช้

Backend: Go (Golang)
Framework: Gin
Database: MySQL
Template: HTML + Bootstrap
Security: Prepared Statement (ป้องกัน SQL Injection)
File Upload: Image Upload ผ่าน Gin

------------------------------------------------------------
🧩 โครงสร้างโปรเจกต์

rental_system/
│
├── main.go                # ไฟล์หลัก
├── templates/             # ไฟล์ HTML Templates
│   ├── index.html
│   ├── dashboard.html
│   ├── add_room.html
│   ├── edit_room.html
│   └── login.html
│
├── uploads/               # โฟลเดอร์เก็บรูปห้องพัก
├── static/                # CSS, JS
└── db.sql                 # สคริปต์สร้างฐานข้อมูล

------------------------------------------------------------
🗄️ โครงสร้างฐานข้อมูล

ตาราง rooms
- id (INT, PK)
- room_number (VARCHAR)
- size (VARCHAR)
- price (FLOAT)
- image (VARCHAR)
- status (VARCHAR)

ตาราง bookings
- id (INT, PK)
- room_id (INT)
- name (VARCHAR)
- phone (VARCHAR)
- start_date (DATE)
- end_date (DATE)
- months (INT)
- status (VARCHAR)

------------------------------------------------------------
🔒 ความปลอดภัย
- ใช้ Prepared Statement สำหรับทุกการ Query
- ตรวจสอบสิทธิ์การเข้าถึงหน้า dashboard และ edit_room
- อัปโหลดรูปภาพแบบจำกัดประเภทไฟล์

------------------------------------------------------------
▶️ วิธีใช้งาน

1. สร้างฐานข้อมูล
นำไฟล์ db.sql ไป Import ลงใน phpMyAdmin หรือรันใน MySQL shell

2. แก้ไขการเชื่อมต่อฐานข้อมูลใน main.go

func connectDB() (*sql.DB, error) {
    return sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/rental_system")
}

3. รันเซิร์ฟเวอร์
go run main.go

ระบบจะรันที่ http://localhost:8080

------------------------------------------------------------
💬 ผู้พัฒนา
Teepakorn Thipphamon
นักศึกษาสาขาเทคโนโลยีสารสนเทศ
โปรเจกต์นี้สร้างขึ้นเพื่อฝึกพัฒนาเว็บด้วย Go และ Gin Framework

------------------------------------------------------------
📜 License
MIT License — สามารถนำไปพัฒนาต่อได้โดยอ้างอิงผู้สร้างเดิม
