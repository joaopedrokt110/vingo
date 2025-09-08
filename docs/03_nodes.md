# 1 - Variables

- Variable node represents dynamic values ​​within the template.

## Example:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title></title>
</head>
<body>
    <p><{ name }></p>
</body>
</html>
```

```go
package main
import (
    "github.com/coderianx/vingo"
    "net/http"
)
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data,err := vingo.Render("template.html", map[string]interface{}{
            "name": "Vingo",
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        
        w.Write([]byte(data))
    })
    http.ListenAndServe(":8080", nil)
}
```
- In this example, the name variable is used to display the value of the name variable.

- The variable node is represented by the syntax `<{ variable_name }>.`

---
# 2 - İf / Else If / Else
- The if node is used to create conditional statements within the template.
## Example:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title></title>
</head>
<body>
    <{ if isLoggedIn }>
        <p>Welcome back, <{ username }>!</p>
    <{ else }>
        <p>Please log in to continue.</p>
    <{ /if }>
</body>
</html>
```

```go
package main

import (
    "github.com/coderianx/vingo"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data,err := vingo.Render("template.html", map[string]interface{}{
            "isLoggedIn": true,
            "username": "VingoUser",
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        
        w.Write([]byte(data))
    })
    http.ListenAndServe(":8080", nil)
}
```
- In this example, the if node checks the value of the isLoggedIn variable. If it is true, it displays a welcome message with the username variable. If it is false, it prompts the user to log in.
- The if node is represented by the syntax `<{ if condition }> ... <{ /if }>.`
- The else if node can be used to check additional conditions, and the else node is used to define a default case when none of the previous conditions are met.
- The else if node is represented by the syntax `<{ else if condition }>.`
- The else node is represented by the syntax `<{ else }>.`
---
# 3 - For Loop
- The for node is used to create loops within the template.
## Example:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Hadi Bee</title>
</head>
<body>
    <ul>
        <{ for item in items }>
            <li><{ item }></li>
        <{ /for }>
    </ul>
</body>
</html>
```

```go
package main
import (
    "github.com/coderianx/vingo"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data,err := vingo.Render("template.html", map[string]interface{}{
            "items": []string{"Item 1", "Item 2", "Item 3"},
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        
        w.Write([]byte(data))
    })
    http.ListenAndServe(":8080", nil)
}
```
- In this example, the for node iterates over the items variable, which is a slice of strings. For each item in the slice, it creates a list item (`<li>`) displaying the value of the item variable.
- The for node is represented by the syntax `<{ for item in collection }> ... <{ /for }>.`
- The item variable represents the current item in the iteration, and the collection variable represents the slice or array being iterated over.
---
# 4 - Switch Case
- The switch node is used to create switch-case statements within the template.
## Example:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Switch Case Example</title>
</head>
<body>
    <{ switch userRole }>
        <{ case "admin" }>
            <p>Welcome, Admin! You have full access.</p>
        <{ case "editor" }>
            <p>Welcome, Editor! You can edit content.</p>
        <{ case "viewer" }>
            <p>Welcome, Viewer! You can view content.</p>
        <{ default }>
            <p>Welcome! Please log in to access more features.</p>
        <{ /switch }>
</body>
</html>
```

```go
package main

import (
    "github.com/coderianx/vingo"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data,err := vingo.Render("template.html", map[string]interface{}{
            "userRole": "editor",
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        
        w.Write([]byte(data))
    })
    http.ListenAndServe(":8080", nil)
}
```
- In this example, the switch node checks the value of the userRole variable. Depending on its value, it displays a different message for "admin", "editor", "viewer", or a default message if none of the cases match.