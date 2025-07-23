package bot

import "math/rand"

var ErrGIFList = []string{
	"https://media1.tenor.com/m/fOY7JqoL7v0AAAAd/bike-fall-memes-baton-roue.gif",
	"https://media1.tenor.com/m/m8EBU4VwPKEAAAAC/bike.gif",
	"https://media1.giphy.com/media/v1.Y2lkPTc5MGI3NjExMW9td25qaGxmMzB0czM2bmU5YzExenhoY3VrdWNob3p4YjRmajVwMiZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/l46Czzp0KEHSO7OdG/giphy.gif",
	"https://media1.tenor.com/m/N8vyVQh1E-gAAAAd/error-loading.gif",
	"https://media1.tenor.com/m/DCI2uoqFUvEAAAAd/the-office-the.gif",
	"https://media1.tenor.com/m/TRe5NGqnQEkAAAAd/were-sorry-tony-hayward.gif",
	"https://media1.tenor.com/m/601zsjqXi-YAAAAd/were-sorry-tony-hayward.gif",
}

func RandomGif() string {
	i := rand.Intn(len(ErrGIFList))
	return ErrGIFList[i]
}
