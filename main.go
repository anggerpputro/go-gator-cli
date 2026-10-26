package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	"github.com/anggerpputro/go-gator-cli/internal"
	"github.com/anggerpputro/go-gator-cli/internal/database"
)

type command struct {
	Name      string
	Arguments []string
}

type commands struct {
	Handlers map[string]func(*internal.State, command) error
}

func (c *commands) run(s *internal.State, cmd command) error {
	callback, ok := c.Handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("can't find the '%s' handler", cmd.Name)
	}
	fmt.Printf("running: %s\n", cmd.Name)
	return callback(s, cmd)
}

func (c *commands) register(name string, f func(*internal.State, command) error) {
	c.Handlers[name] = f
}

func main() {
	state, err := internal.NewState()
	if err != nil {
		log.Fatalf("error creating new state: %v", err)
	}

	cmds := commands{
		Handlers: map[string]func(*internal.State, command) error{},
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		// fmt.Fprintf(os.Stderr, "")
		log.Fatal("args must be larger than 2")
	}
	// fmt.Printf("args: %+v\n", args)

	if err := cmds.run(state, command{
		Name:      args[1],
		Arguments: args[2:],
	}); err != nil {
		log.Fatal(err)
	}

	// fmt.Printf(".gatorconfig.json: %+v\n", state.Config)
}

func middlewareLoggedIn(handler func(s *internal.State, cmd command, user database.User) error) func(*internal.State, command) error {
	return func(s *internal.State, cmd command) error {
		username := s.Config.CurrentUserName
		if username == "" {
			return errors.New("please log in first")
		}

		user, err := s.DB.GetUser(context.Background(), username)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("current user %s no longer exists; please log in again", username)
		}
		if err != nil {
			// return fmt.Errorf("[ADDFEED] ERROR! User login: '%s' not found!", s.Config.CurrentUserName)
			return fmt.Errorf("get current user: %w", err)
		}

		return handler(s, cmd, user)
	}
}

func handlerLogin(s *internal.State, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("[LOGIN] ERROR! Usage: login [username]")
	}

	// username := "anggerpputro"
	ctx := context.Background()
	username := cmd.Arguments[0]

	userExists, err := s.DB.CheckUserExists(ctx, username)
	if err != nil {
		return err
	}
	if !userExists {
		return fmt.Errorf("[LOGIN] ERROR! User with username: '%s' doesn't exists!", username)
	}

	if err := s.Config.SetUser(username); err != nil {
		return fmt.Errorf("error set config username: %w", err)
	}

	fmt.Printf("[LOGIN] SUCCESS! Username set to: %s\n", username)

	return nil
}

func handlerRegister(s *internal.State, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("[REGISTER] ERROR! Usage: register [username]")
	}

	ctx := context.Background()
	username := cmd.Arguments[0]

	userExists, err := s.DB.CheckUserExists(ctx, username)
	if err != nil {
		return err
	}
	if userExists {
		return fmt.Errorf("[REGISTER] ERROR! User with username: '%s' already exists!", username)
	}

	newUser, err := s.DB.CreateUser(ctx, username)
	if err != nil {
		return err
	}

	if err := s.Config.SetUser(newUser.Name); err != nil {
		return fmt.Errorf("error set config username: %w", err)
	}

	fmt.Printf("[REGISTER] SUCCESS! User registered: %+v\n", newUser)

	return nil
}

func handlerReset(s *internal.State, cmd command) error {
	ctx := context.Background()
	if err := s.DB.DeleteAllUsers(ctx); err != nil {
		return err
	}

	fmt.Println("[RESET] SUCCESS! Reset users table!")

	return nil
}

func handlerUsers(s *internal.State, cmd command) error {
	ctx := context.Background()

	users, err := s.DB.ListUsers(ctx)
	if err != nil {
		return err
	}

	currentUsername := s.Config.CurrentUserName
	fmt.Println("[USERS] SUCCESS! List all users:")

	for _, v := range users {
		print := fmt.Sprintf("* %s", v.Name)
		if v.Name == currentUsername {
			print = fmt.Sprintf("%s (current)", print)
		}
		fmt.Printf("%s\n", print)
	}

	return nil
}

func handlerAgg(s *internal.State, cmd command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("[AGG] ERROR! Usage: agg [time_between_reqs]\nExample: agg 1m")
	}

	timeBeetweenReqs, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", timeBeetweenReqs)

	// ctx := context.Background()

	// feedURL := "https://www.wagslane.dev/index.xml"
	// rss, err := internal.FetchFeed(ctx, feedURL)
	// if err != nil {
	// 	return fmt.Errorf("[AGG] ERROR! %w", err)
	// }

	// fmt.Println("[AGG] SUCCESS! Feed fetched:")
	// for _, i := range rss.Channel.Item {
	// 	fmt.Printf("%s\n", i.Title)
	// 	fmt.Printf("%s\n", i.Description)
	// 	fmt.Println("===================================")
	// }

	ticker := time.NewTicker(timeBeetweenReqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			ticker.Stop()
			return err
		}
	}
}

func handlerAddFeed(s *internal.State, cmd command, userLogin database.User) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("[ADDFEED] ERROR! Usage: addfeed [name] [url]")
	}

	ctx := context.Background()

	// userLogin, err := s.DB.GetUser(ctx, s.Config.CurrentUserName)
	// if err != nil {
	// 	return fmt.Errorf("[ADDFEED] ERROR! User login: '%s' not found!", s.Config.CurrentUserName)
	// }

	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]

	feedCreated, err := s.DB.CreateFeed(ctx, database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: userLogin.ID,
	})
	if err != nil {
		return err
	}

	feedFollowCreated, err := s.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: userLogin.ID,
		FeedID: feedCreated.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("[ADDFEED] SUCCESS! Feed created: %+v\n", feedCreated)
	fmt.Printf("[ADDFEED] SUCCESS! Feed Follow created: %+v\n", feedFollowCreated)

	return nil
}

func handlerFeeds(s *internal.State, cmd command) error {
	ctx := context.Background()

	feeds, err := s.DB.ListFeeds(ctx)
	if err != nil {
		return err
	}

	fmt.Println("[FEEDS] SUCCESS! List Feeds:")
	for _, f := range feeds {
		fmt.Printf("- Feed Name: %s\n", f.Name)
		fmt.Printf("- Feed URL: %s\n", f.Url)
		fmt.Printf("- User Name: %s\n", f.UserName)
		fmt.Println("=====================================")
	}

	return nil
}

func handlerFollow(s *internal.State, cmd command, userLogin database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("[FOLLOW] ERROR! Usage: follow [url]")
	}
	url := cmd.Arguments[0]

	ctx := context.Background()
	// userNameLogin := s.Config.CurrentUserName

	// user, err := s.DB.GetUser(ctx, userNameLogin)
	// if err != nil {
	// 	return err
	// }
	// if !isUserExists {
	// 	return fmt.Errorf("[FOLLOW] ERROR! user with username: '%s' doesn't exists!", userNameLogin)
	// }

	feed, err := s.DB.GetFeed(ctx, url)
	if err != nil {
		return err
	}
	// if !isFeedExists {
	// 	return fmt.Errorf("[FOLLOW] ERROR! feed with url: '%s' doesn't exists!", url)
	// }

	feedFollowCreated, err := s.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: userLogin.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println("[FOLLOW] SUCCESS! Feed followed!")
	fmt.Printf("- URL: %s\n", url)
	fmt.Printf("- Feed Name: %s\n", feedFollowCreated.FeedName)
	fmt.Printf("- Current User Name: %s\n", feedFollowCreated.UserName)

	return nil
}

func handlerFollowing(s *internal.State, cmd command, userLogin database.User) error {
	ctx := context.Background()
	// userNameLogin := s.Config.CurrentUserName
	userNameLogin := userLogin.Name

	listFeedFollowed, err := s.DB.GetFeedFollowsForUser(ctx, userNameLogin)
	if err != nil {
		return err
	}

	fmt.Printf("[FEEDS] SUCCESS! List Feeds Followed (username: '%s'):\n", userNameLogin)
	for _, f := range listFeedFollowed {
		fmt.Printf("- %s\n", f.FeedName)
	}

	return nil
}

func handlerUnfollow(s *internal.State, cmd command, userLogin database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("[UNFOLLOW] ERROR! Usage: unfollow [feed_url]")
	}

	ctx := context.Background()
	feedUrl := cmd.Arguments[0]

	feedData, err := s.DB.GetFeed(ctx, feedUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("[UNFOLLOW] ERROR! Can't find feed with url: '%s'", feedUrl)
	}
	if err != nil {
		return err
	}

	if err := s.DB.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: userLogin.ID,
		FeedID: feedData.ID,
	}); err != nil {
		return err
	}

	fmt.Println("[UNFOLLOW] SUCCESS! Feed unfollowed!")

	return nil
}

func scrapeFeeds(s *internal.State) error {
	fmt.Println("[SCRAPE-FEEDS] Called")

	ctx := context.Background()

	feedToScrape, err := s.DB.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("[SCRAPE-FEEDS] Scraping: %s\n", feedToScrape.Url)

	if err := s.DB.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: feedToScrape.ID,
	}); err != nil {
		return err
	}

	rss, err := internal.FetchFeed(ctx, feedToScrape.Url)
	if err != nil {
		return err
	}

	fmt.Printf("[SCRAPE-FEEDS] SUCCESS! Feed fetched: (url: %s)\n", feedToScrape.Url)
	for k, i := range rss.Channel.Item {
		// fmt.Printf("[%d] %s\n", k, i.Title)
		// fmt.Printf("%s\n", i.Link)
		// fmt.Printf("%s\n", i.Description)
		// fmt.Printf("%s\n", i.PubDate)

		parsedPubDate, err := internal.TimeParser(i.PubDate)
		if err != nil {
			return err
		}
		postExists, err := s.DB.CheckPostExists(ctx, i.Link)
		if err != nil {
			return err
		}
		if !postExists {
			// post not exists, create
			postCreated, err := s.DB.CreatePost(ctx, database.CreatePostParams{
				Title: i.Title,
				Url:   i.Link,
				Description: sql.NullString{
					String: i.Description,
					Valid:  true,
				},
				PublishedAt: sql.NullTime{
					Time:  parsedPubDate,
					Valid: true,
				},
				FeedID: feedToScrape.ID,
			})
			if err != nil {
				return err
			}
			fmt.Printf("[%d] %s\n", k, postCreated.Title)
		} else {
			// post exists, update
			postUpdated, err := s.DB.UpdatePost(ctx, database.UpdatePostParams{
				Title: i.Title,
				Url:   i.Link,
				Description: sql.NullString{
					String: i.Description,
					Valid:  true,
				},
				PublishedAt: sql.NullTime{
					Time:  parsedPubDate,
					Valid: true,
				},
				FeedID:   feedToScrape.ID,
				WhereUrl: i.Link,
			})
			if err != nil {
				return err
			}
			fmt.Printf("[%d] %s\n", k, postUpdated.Title)
		}
	}
	fmt.Println("===================================")

	return nil
}

func handlerBrowse(s *internal.State, cmd command, userLogin database.User) error {
	limit := 2
	if len(cmd.Arguments) == 1 {
		num, err := strconv.Atoi(cmd.Arguments[0])
		if err != nil {
			return err
		}
		limit = num
	}

	ctx := context.Background()
	userPosts, err := s.DB.GetUserPosts(ctx, database.GetUserPostsParams{
		UserID: userLogin.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return err
	}

	fmt.Printf("[BROWSE] SUCCESS! Posts by user: (name: %s)\n", userLogin.Name)
	for k, p := range userPosts {
		fmt.Printf("[%d] post url: %s\n", k, p.Url)
		fmt.Printf("  - post title: %s\n", p.Title)
		fmt.Printf("  - feed name: %s\n\n", p.FeedName)
	}
	fmt.Println("===================================")

	return nil
}
