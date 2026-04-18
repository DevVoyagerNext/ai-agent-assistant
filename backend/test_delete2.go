package main

import (
"fmt"
"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

type Message struct {
ID      int64  gorm:"primaryKey;autoIncrement"
Content string
}

func main() {
db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
if err != nil {
panic(err)
}
db.AutoMigrate(&Message{})

msg1 := Message{Content: "msg1"}
msg2 := Message{Content: "msg2"}
db.Create(&msg1)
db.Create(&msg2)

var msgs []Message
db.Find(&msgs)
fmt.Println("Before:", len(msgs))

res := db.Unscoped().Where("id IN ?", []int64{msg1.ID, msg2.ID}).Delete(&Message{})
fmt.Println("Deleted:", res.RowsAffected)

db.Find(&msgs)
fmt.Println("After:", len(msgs))
}
