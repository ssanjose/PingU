package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ssanjose/PingU/internal/store"
)

var usernames = []string{
	"alice", "bob", "dave", "eve", "frank", "grace", "hank", "iris", "jack", "kate",
	"leo", "mike", "nancy", "oliver", "peter", "quinn", "rachel", "sam", "tina", "ursula",
	"vick", "walt", "xander", "yara", "zane", "anna", "bill", "claire", "dan", "ella",
	"fred", "gina", "hannah", "ian", "jane", "ken", "linda", "mary", "noah", "oscar",
	"paul", "quincy", "rose", "sophie", "tom", "van", "will", "xena", "yuki",
}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating user:", err)
			return
		}
	}

	tx.Commit()

	log.Println("Seeding complete")
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
		}
	}

	return users
}
