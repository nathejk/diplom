package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/httpimg"
	"golang.org/x/text/encoding/charmap"
	"nathejk.dk/internal/data"
)

type Year struct {
	Title       string
	Start       string
	Destination string
	Background  string
}

type Team struct {
	Name       string
	Number     string
	FinishedAt time.Time
	PhotoUrl   string
}

func (app *application) pdfHandler(w http.ResponseWriter, r *http.Request) {
	teamNumber := app.ReadNamedParam(r, "id")
	if teamNumber == "" {
		app.NotFoundResponse(w, r)
		return
	}
	year := Year{
		Title:       "Nathejk 2024",
		Start:       "Lundby",
		Destination: "Glumsø",
		Background:  "/app/assets/Diplom2024_patrulje.png",
	}
	//loc, _ := time.LoadLocation("Europe/Copenhagen")
	/*team := Team{
		Name:   "Skovskiderne",
		Number: "123-4",
		//FinishedAt: time.Now().In(loc),
		PhotoUrl: "https://natpas.nathejk.dk/photo.image.php?id=1688",
	}*/
	team, err := app.models.Teams.GetPatruljeByYearAndNumber("2024", teamNumber)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}
	photoUrl := fmt.Sprintf("https://natpas.nathejk.dk/photo.image.php?id=%s", team.PhotoID)
	switch teamNumber {
	case "81":
		photoUrl = "/app/assets/Team-81-3_1_XXX.jpg"
	case "107":
		photoUrl = "/app/assets/Team-107-5_1_XXX.jpg"
	}
	checkpoint, _ := app.models.Checkpoint.GetLastCheckpoint("2024")
	var finishedAt time.Time
	if scan, err := app.models.Scan.GetByTeamIDAndCheckpoint(team.ID, checkpoint.GroupID); err == nil {
		finishedAt = time.Unix(scan.Uts, 0)
	}
	pdf := fpdf.New("P", "mm", "A4", "/")
	pdf.AddPage()

	//httpimg.Register(pdf, year.Background, "")
	pdf.Image(year.Background, 0, 0, 210, 297, false, "", 0, "")

	pdf.AddUTF8Font("impact", "", "app/assets/impact.ttf")

	pdf.SetFont("impact", "", 20)
	pdf.SetXY(10, 210)
	pdf.MultiCell(0, 10, team.Name, "", "C", false)

	if team.PhotoID != "" {
		if photoUrl[0:4] == "http" {
			httpimg.Register(pdf, photoUrl, "")
		}
		pdf.Image(photoUrl, 55, 130, 100, 75, false, "", 0, "")
	}
	pdf.SetXY(65, 220)
	pdf.SetFont("Arial", "", 12)
	if !finishedAt.IsZero() {
		//ts := strftime('%R', $team->finishUts);
		pdf.MultiCell(80, 5, utf8_decode(fmt.Sprintf("har gennemført %s", year.Title)), "", "C", false)
		pdf.SetX(65)
		pdf.MultiCell(80, 5, utf8_decode(fmt.Sprintf("fra %s til %s", year.Start, year.Destination)), "", "C", false)
		pdf.SetX(65)
		pdf.MultiCell(80, 5, utf8_decode(fmt.Sprintf("og gik i mål lørdag nat kl. %s!", finishedAt.Format("15:04"))), "", "C", false)
	} else {
		text := fmt.Sprintf("deltog i %s fra %s til %s!", year.Title, year.Start, year.Destination)
		pdf.Cell(80, 5, utf8_decode(text))
	}

	err = pdf.Output(w)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func utf8_decode(s string) string {
	encoder := charmap.ISO8859_1.NewEncoder()
	out, _ := encoder.Bytes([]byte(s))
	return string(out)
}
