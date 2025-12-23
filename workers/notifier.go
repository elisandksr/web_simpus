package workers

import (
	"fmt"
	"latihan_cloud8/store"
	"log"
	"time"
)

type Notifier struct {
	Store *store.MySQLStore
}

func NewNotifier(store *store.MySQLStore) *Notifier {
	return &Notifier{Store: store}
}

func (n *Notifier) Start() {
	ticker := time.NewTicker(24 * time.Hour)
	// Run immediately on start (or delay)
	go func() {
		n.Check() // Initial check
		for range ticker.C {
			n.Check()
		}
	}()
}

func (n *Notifier) Check() {
	log.Println("Worker: Checking for overdue books and reminders...")
	loans, err := n.Store.GetAllBorrowedLoans()
	if err != nil {
		log.Println("Worker Error:", err)
		return
	}

	settings, _ := n.Store.GetSettings()
	finePerDay := 5000
	if settings != nil {
		finePerDay = settings.FinePerDay
	}

	for _, l := range loans {
		book, _ := n.Store.GetBookByID(l.BookID)
		bookTitle := "Buku"
		if book != nil {
			bookTitle = book.Title
		}

		now := time.Now()
		// Check Overdue
		if now.After(l.DueDate) {
			daysLate := int(now.Sub(l.DueDate).Hours() / 24)
			if daysLate < 1 {
				daysLate = 1
			} // Minimum 1 day if passed due date
			fine := daysLate * finePerDay

			msg := fmt.Sprintf("PERINGATAN: Buku '%s' terlambat %d hari. Denda sementara: Rp %d. Segera kembalikan!", bookTitle, daysLate, fine)
			// Avoid spamming? Ideally yes, but per requirement "reminder notification".
			// We'll send it. User might get one every day until returned.
			n.Store.CreateNotification(l.UserID, msg)
		}

		// Check Reminder (1 Day Before)
		// DueDate - Now < 24h && DueDate > Now
		durationUntilDue := l.DueDate.Sub(now)
		if durationUntilDue > 0 && durationUntilDue < 24*time.Hour {
			msg := fmt.Sprintf("PENGINGAT: Buku '%s' harus dikembalikan besok (%s).", bookTitle, l.DueDate.Format("02 Jan 2006"))
			n.Store.CreateNotification(l.UserID, msg)
		}
	}
}
