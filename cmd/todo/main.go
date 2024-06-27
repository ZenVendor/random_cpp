package main

import (
	"fmt"
	"log"
	"time"

    "./internal/functions"
)

const VERSION = "0.6.2"

func main () {
    var conf functions.Config
    err := conf.Prepare()
    if err != nil {
        log.Fatal(err)
    }

    cmd, sw, vals, valid := functions.ParseArgs(conf.dateFormat)

    if !valid {
        PrintVersion()
        PrintHelp()
        return
    }
    if cmd == functions.CMD_VERSION {
        PrintVersion()
        return
    }

    db, err := conf.OpenDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if cmd == functions.CMD_COUNT {
        count, err := functions.Count(db, sw)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%d", count)
        return
    }

    if cmd == functions.CMD_COMPLETE {
        taskId := vals.ReadValue("id").(int)
        err := functions.Complete(db, taskId)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Task %d has been completed\n", taskId)
    }

    if cmd == functions.CMD_REOPEN {
        taskId := vals.ReadValue("id").(int)
        err := functions.Reopen(db, taskId)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Task %d has been reopened\n", taskId)
    }

    if cmd == functions.CMD_DELETE {
        taskId := vals.ReadValue("id").(int)
        err := functions.Delete(db, taskId)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Task %d has been deleted\n", taskId)
    }

    if cmd == functions.CMD_ADD {
        var t functions.Task
        t.description = vals.ReadValue("description").(string)
        if vals.ReadValue("due") != nil {
            t.due = vals.ReadValue("due").(time.Time)
        }      
        t.created = time.Now()
        t.updated = time.Now()

        if err = t.AddTask(db); err != nil {
            log.Fatal(err)
        }

        fmt.Printf("Added task: %s\n", t.description)
        if t.due.Year() != 1 {
            fmt.Printf("Due date: %s\n", t.due.Format(conf.dateFormat))
        }
    }
   
    if cmd == functions.CMD_LIST {
        count, err := functions.Count(db, sw)
        if err != nil {
            log.Fatal(err)
        }
        tl, err := functions.List(db, sw)
        if err != nil {
            log.Fatal(err)
        }
        var tType string 
        switch sw {
            case functions.SW_OPEN:
                tType = "Open"
            case functions.SW_CLOSED:
                tType = "Closed"
            case functions.SW_ALL:
                tType = "All"
            case functions.SW_OVERDUE:
               tType = "Overdue"
        }
        fmt.Printf("%s tasks: %d\n", tType, count)

        for _, t := range tl {
            tStatus := "Open"
            if t.done == 0 && t.due.Year() != 1 && t.due.Before(time.Now()) {
                tStatus = "Overdue"
            }
            if t.done == 1 {
                tStatus = "Closed"
            }
            if t.done == 0 {
                if t.due.Year() == 1 {
                    fmt.Printf("\t%s %d: %s\n", tStatus, t.id, t.description)
                } else {
                    fmt.Printf("\t%s %d: %s, due date: %s\n", tStatus, t.id, t.description, t.due.Format(conf.dateFormat))
                }
            } else {
                fmt.Printf("\t%s %d: %s, completed: %s\n", tStatus, t.id, t.description, t.completed.Format(conf.dateFormat))
            }
        }
        fmt.Printf("End.\n")
    }
    
    if cmd == functions.CMD_UPDATE {
        var t functions.Task

        taskId := vals.ReadValue("id").(int)
        t, err = functions.Select(db, taskId)
        if err != nil {
            log.Fatal(err)
        }

        if vals.ReadValue("description") != nil {
            t.description = vals.ReadValue("description").(string)
        }
        if vals.ReadValue("due") != nil {
            t.due = vals.ReadValue("due").(time.Time)
        }
        t.updated = time.Now()
        if err = t.Update(db); err != nil {
            log.Fatal(err)
        }
        
        var updString string
        for i, v := range vals {
            if i != 0 {
                updString += ", "
            }
            updString += v.name
        }
        fmt.Printf("Updated %s in task %d\n", updString, t.id)
    }
    return
}

func PrintVersion() {
    fmt.Printf("TODO CLI\tversion: %s\n", VERSION)
}

func PrintHelp() {
    helpString := `
Usage: 
    todo [command] [id] [option] [argument]

Without arguments defaults to listing active tasks.
Frequently used commands have single-letter aliases.
In ADD command, description is required and must be provided first.
In commands that require it, task ID must follow the command.
Values following switches can be provided in any order.

    
    help | h | --help | -h                      display this help

    version | v | --version | -v                display program version

    add | a [description] [due]                 optional due date format 2006-01-02

    count                                       defaults to active tasks
        --completed | -c
        --overdue | -o
        --all | -a

    list | l                                    defaults to active tasks
        --completed | -c
        --overdue | -o
        --all | -a
        
    update | u [id]                             update description, due date, or both. invalid date value removes due date
        --desc [description] 
        --due [date]

    complete | c [task_id]                      set task completed

    reopen | open [task_id]                     reopen completed task

    delete | del [task_id]                      delete task

Examples:
    todo
    todo a "New task"
    todo add "New task" "2024-08-13"
    todo list --all
    todo l -o
    todo count -c
    todo update 15 --due "2024-08-13"
    todo u 10 --due -
    todo c 12
    todo reopen 3
    todo del 5

`
    fmt.Println(helpString)
}
