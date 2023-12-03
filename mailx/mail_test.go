package mailx

import (
	"testing"
)

func TestFetchNewOrders(t *testing.T) {
	SendMailHtmlWithAttach(MailClassReport, "wefwkeelkiiu", "postgres", []string{"chart.png", "chart.png"})
}
