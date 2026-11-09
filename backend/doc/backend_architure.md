สถาปัตยกรรม Backend

ภาพรวม

Backend ใช้สถาปัตยกรรมแบบแบ่งชั้น (Layered Architecture) เพื่อแยกหน้าที่และความรับผิดชอบของแต่ละส่วนออกจากกัน ช่วยให้โค้ดเป็นระเบียบ ง่ายต่อการบำรุงรักษา และรองรับการขยายระบบในอนาคต

1. Main

หน้าที่: จุดเริ่มต้นและเตรียมความพร้อมของระบบ (Application Entry Point and Initialization)

โหลดการตั้งค่าของระบบ (Configuration)

อ่านค่าตัวแปรสภาพแวดล้อม (Environment Variables)

เชื่อมต่อฐานข้อมูล (Database Connection)

สร้างและกำหนดการทำงานร่วมกันของแต่ละส่วน (Dependency Initialization)

กำหนดเส้นทาง API (Route Registration)

กำหนด Middleware

เริ่มต้น HTTP Server

2. Handler

หน้าที่: จัดการคำขอและการตอบกลับ HTTP (HTTP Request and Response Handling)

รับคำขอ HTTP จาก Client

อ่านและตรวจสอบข้อมูลที่ส่งเข้ามา (Request Validation)

เรียกใช้งาน Service ที่เกี่ยวข้อง

ส่งผลลัพธ์กลับไปยัง Client

กำหนด HTTP Status Code

จัดการข้อผิดพลาดในระดับ HTTP

3. Service

หน้าที่: จัดการตรรกะและเงื่อนไขการทำงานของระบบ (Business Logic)

ประมวลผลตามเงื่อนไขทางธุรกิจ

ตรวจสอบความถูกต้องของข้อมูลตามกฎของระบบ

ประสานการทำงานระหว่าง Repository

จัดการลำดับขั้นตอนการทำงานของแต่ละ Use Case

จัดการข้อผิดพลาดในระดับ Business Logic

4. Repository

หน้าที่: จัดการการเข้าถึงและจัดเก็บข้อมูล (Data Access and Persistence)

ติดต่อกับฐานข้อมูล (Database)

ดำเนินการคำสั่ง SQL

เพิ่ม แก้ไข ลบ และค้นหาข้อมูล (CRUD Operations)

แปลงข้อมูลระหว่างฐานข้อมูลกับ Model

จัดการข้อผิดพลาดที่เกิดจากฐานข้อมูล

ลำดับการทำงาน

Client → Handler → Service → Repository → Database

เมื่อประมวลผลเสร็จสิ้น ผลลัพธ์จะถูกส่งกลับตามลำดับย้อนกลับไปยัง Client

หลักการออกแบบ

แยกหน้าที่ของแต่ละส่วนให้ชัดเจน (Separation of Concerns)

แต่ละส่วนรับผิดชอบงานของตนเอง (Single Responsibility Principle)

ลดการพึ่งพาระหว่างส่วนต่าง ๆ (Loose Coupling)

สามารถทดสอบแต่ละส่วนแยกจากกันได้ (Testability)

รองรับการบำรุงรักษาและขยายระบบในอนาคต (Maintainability and Scalability)