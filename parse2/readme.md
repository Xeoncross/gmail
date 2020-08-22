# multipart/mime email reader

Clean, in-memory version
- https://github.com/lavab/mailer/blob/master/handler/parser.go
Buffer only when needed (saves disk I/O)
- https://www.reddit.com/r/golang/comments/8bu8m3/buffered_iowriter_switching_to_filebacking_for/
