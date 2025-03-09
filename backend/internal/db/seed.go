package db

import (
	"context"
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

func Seed(store store.Storage) {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Error creating user:", err)
			return
		}
	}

	log.Println("Seeding complete")
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Password: "123123",
		}
	}

	return users
}
