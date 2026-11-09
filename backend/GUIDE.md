Architecture
Folder	หน้าที่
cmd/api	จุดเริ่มต้น Backend
internal/config	ตั้งค่า Environment
internal/handler	รับ HTTP Request
internal/service	Business Logic
internal/repository	ติดต่อ Database
internal/model	โครงสร้างข้อมูล
migrations	คำสั่งสร้างและแก้ไข Database


go mod init ${path} สร้าง Go Module

go get github.com/shirou/gopsutil/v4  ติดตั้ง Library

go mod tidy ตรวจสอบ


go fmt -w internal/handler/server_handler.go  คำสั่งของ Go สำหรับ จัดรูปแบบโค้ดอัตโนมัติ (Code Formatting)
go build ./... คำสั่งสำหรับ Compile โค้ด Go ทั้งโปรเจกต์ เพื่อตรวจสอบว่าโค้ดสามารถ Build ได้หรือไม่




#คำสั่งนี้สร้าง Environment Variable ชื่อ DATABASE_URL เพื่อบอก Backend ว่าจะเชื่อมต่อฐานข้อมูลที่ไหน

ส่วน	ความหมาย
postgres://	Protocol สำหรับ PostgreSQL
dashboard_user	Username
change_this_password	Password
localhost	Database อยู่ในเครื่องเรา
5432	Port ของ PostgreSQL
home_dashboard	ชื่อ Database
sslmode=disable	ไม่ใช้ SSL สำหรับการเชื่อมต่อ Local นี้

$env:DATABASE_URL="postgres://dashboard_user:change_this_password@localhost:5432/home_dashboard?sslmode=disable"


#Compile โปรแกรม Go รันโปรแกรมทันที
-go run ./cmd/api

http://localhost:8080/health


Method	Endpoint	หน้าที่
GET	/health	ตรวจสอบ Backend
GET	/api/v1/servers	อ่าน Server
POST	/api/v1/servers	เพิ่ม Server
DELETE	/api/v1/servers/{server_id}	ลบ Server ของผู้ใช้ที่ล็อกอิน
POST	/api/v1/metrics	รับข้อมูล Hardware
GET	/api/v1/servers/{server_id}/metrics	อ่าน Metrics ย้อนหลัง
GET	/api/v1/servers/{server_id}/metrics	อ่าน Metrics ย้อนหลัง


ระบบ	ตรวจสอบอะไร
Authentication	คุณเป็นใคร?
Authorization	คุณมีสิทธิ์เข้าถึงอะไร?
