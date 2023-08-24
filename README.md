# SnippetBox

[![Go Version](https://img.shields.io/github/go-mod/go-version/anshul393/snippetbox)](https://golang.org/doc/go1.20)

SnippetBox is a web application built in Go that allows users to store, manage, and share code snippets easily. It utilizes various third-party libraries for session management, routing, and more.

## Features

- Store and manage your code snippets in one place.
- Secure session management using [SCS](https://github.com/alexedwards/scs) and Redis as a store.
- User-friendly routes handling powered by [HttpRouter](https://github.com/julienschmidt/httprouter).
- Input validation and security with libraries like [nosurf](https://github.com/justinas/nosurf) for CSRF protection.

## Installation and Usage

Follow these steps to set up and run SnippetBox on your local machine:

1. Clone the repository:

   ```bash
   git clone https://github.com/anshul393/snippetbox.git
   cd snippetbox
