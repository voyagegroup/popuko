package main

import (
	"log"
	"strings"
)

// XXX: parseCommand is doing adhoc command parsing.
// for the future, we should write an actual parser.
func parseCommand(raw string) (ok bool, cmd interface{}) {
	log.Printf("debug: input: %v\n", raw)
	tmp := strings.Split(raw, "\n")

	// If there are no possibility that the comment body is not formatted
	// `@botname command`, stop to process.
	if len(tmp) < 1 {
		return false, nil
	}

	body := tmp[0]
	log.Printf("debug: body: %v\n", body)

	command := strings.Split(body, " ")
	log.Printf("debug: command: %#v\n", command)
	if len(command) < 2 {
		return false, nil
	}

	trigger := command[0]
	if strings.Index(trigger, "@") != 0 {
		return false, nil
	}

	args := strings.Trim(command[1], " ")
	log.Printf("debug: trigger: %v\n", trigger)
	log.Printf("debug: args: %#v\n", args)

	if args == "r?" {
		return true, &AssignReviewerCommand{
			Reviewer: strings.TrimPrefix(trigger, "@"),
		}
	}

	if args == "r+" {
		return true, &AcceptChangeByReviewerCommand{
			BotName: strings.TrimPrefix(trigger, "@"),
		}
	}

	if strings.Index(args, "r=") != 0 {
		return false, nil
	}

	args = strings.TrimPrefix(args, "r=")
	log.Printf("debug: args: %#v\n", args)
	reviwers := strings.Split(args, ",")
	log.Printf("debug: reviwers: %#v\n", reviwers)

	for i, name := range reviwers {
		reviwers[i] = strings.Trim(name, " ")
	}

	return true, &AcceptChangeByOthersCommand{
		BotName:  strings.TrimPrefix(trigger, "@"),
		Reviewer: reviwers,
	}
}

type AcceptChangeByReviewerCommand struct {
	BotName string
}

type AcceptChangeByOthersCommand struct {
	BotName  string
	Reviewer []string
}

type AssignReviewerCommand struct {
	Reviewer string
}
