/*
 * This example showcases PDF report generation with unidoc's creator package.
 * The output is saved as unidoc-report.pdf which illustrates some of the features
 * of the creator.
 */
/*
 * NOTE: This example depends on github.com/boombuler/barcode, MIT licensed,
 *       and github.com/wcharczuk/go-chart, MIT licensed,
 *       and the Roboto font (Roboto-Bold.ttf, Roboto-Regular.ttf), Apache-2 licensed.
 */

package main

import (
	"bytes"
	"fmt"
	goimage "image"
	"math"
	"time"

	"github.com/wcharczuk/go-chart"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	unicommon "github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/creator"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func main() {
	// For development:
	//unicommon.SetLogger(unicommon.NewConsoleLogger(unicommon.LogLevelDebug))

	err := RunPdfReport("unidoc-report.pdf")
	if err != nil {
		panic(err)
	}
}

func RunPdfReport(outputPath string) error {
	robotoFontRegular, err := model.NewPdfFontFromTTFFile("./Roboto-Regular.ttf")
	if err != nil {
		return err
	}

	robotoFontPro, err := model.NewPdfFontFromTTFFile("./Roboto-Bold.ttf")
	if err != nil {
		return err
	}

	c := creator.New()
	c.SetPageMargins(50, 50, 100, 70)

	logoImg, err := creator.NewImageFromFile("./unidoc-logo.png")
	if err != nil {
		return err
	}

	logoImg.ScaleToHeight(25)
	logoImg.SetPos(58, 20)

	DoDocumentControl(c, robotoFontRegular, robotoFontPro)

	DoFeatureOverview(c, robotoFontRegular, robotoFontPro)

	// Setup a front page (always placed first).
	c.CreateFrontPage(func(args creator.FrontpageFunctionArgs) {
		DoFirstPage(c, robotoFontRegular, robotoFontPro)
	})

	// Draw a header on each page.
	c.DrawHeader(func(block *creator.Block, args creator.HeaderFunctionArgs) {
		// Draw the header on a block. The block size is the size of the page's top margins.
		block.Draw(logoImg)
	})

	// Draw footer on each page.
	c.DrawFooter(func(block *creator.Block, args creator.FooterFunctionArgs) {
		// Draw the on a block for each page.
		p := creator.NewParagraph("unidoc.io")
		p.SetFont(robotoFontRegular)
		p.SetFontSize(8)
		p.SetPos(50, 20)
		p.SetColor(creator.ColorRGBFrom8bit(63, 68, 76))
		block.Draw(p)

		strPage := fmt.Sprintf("Page %d of %d", args.PageNum, args.TotalPages)
		p = creator.NewParagraph(strPage)
		p.SetFont(robotoFontRegular)
		p.SetFontSize(8)
		p.SetPos(300, 20)
		p.SetColor(creator.ColorRGBFrom8bit(63, 68, 76))
		block.Draw(p)
	})

	// Generate the table of contents.
	c.CreateTableOfContents(func(toc *creator.TableOfContents) (*creator.Chapter, error) {
		ch := c.NewChapter("Table of contents")
		ch.GetHeading().SetFontSize(28)
		ch.GetHeading().SetMargins(0, 0, 0, 30)

		// Arrange the TOC in a table with first column representing chapter number and title.
		// and second column representing page number.
		table := creator.NewTable(2)
		table.SetColumnWidths(0.9, 0.1)

		for _, entry := range toc.Entries() {
			// Col 1. Chapter number, title.
			var str string
			if entry.Subchapter == 0 {
				str = fmt.Sprintf("%d. %s", entry.Chapter, entry.Title)
			} else {
				str = fmt.Sprintf("        %d.%d. %s", entry.Chapter, entry.Subchapter, entry.Title)
			}

			p := creator.NewParagraph(str)
			p.SetFontSize(14)
			cell := table.NewCell()
			cell.SetContent(p)

			// Col 1. Page number.
			p = creator.NewParagraph(fmt.Sprintf("%d", entry.PageNumber))
			p.SetFontSize(14)
			cell = table.NewCell()
			cell.SetContent(p)
		}
		err := ch.Add(table)
		if err != nil {
			fmt.Printf("Error adding table: %v\n", err)
			return nil, err
		}

		return ch, nil
	})

	err = c.WriteToFile(outputPath)
	if err != nil {
		return err
	}

	return nil
}

// Generates the front page.
func DoFirstPage(c *creator.Creator, fontRegular *model.PdfFont, fontBold *model.PdfFont) {
	helvetica := fonts.NewFontHelvetica()
	helveticaBold := fonts.NewFontHelveticaBold()

	p := creator.NewParagraph("UniDoc")
	p.SetFont(helvetica)
	p.SetFontSize(48)
	p.SetMargins(85, 0, 150, 0)
	p.SetColor(creator.ColorRGBFrom8bit(56, 68, 77))
	c.Draw(p)

	p = creator.NewParagraph("Example Report")
	p.SetFont(helveticaBold)
	p.SetFontSize(30)
	p.SetMargins(85, 0, 0, 0)
	p.SetColor(creator.ColorRGBFrom8bit(45, 148, 215))
	c.Draw(p)

	t := time.Now().UTC()
	dateStr := t.Format("1 Jan, 2006 15:04")

	p = creator.NewParagraph(dateStr)
	p.SetFont(helveticaBold)
	p.SetFontSize(12)
	p.SetMargins(90, 0, 5, 0)
	p.SetColor(creator.ColorRGBFrom8bit(56, 68, 77))
	c.Draw(p)
}

// Document control page.
func DoDocumentControl(c *creator.Creator, fontRegular *model.PdfFont, fontBold *model.PdfFont) {
	ch := c.NewChapter("Document control")
	ch.SetMargins(0, 0, 40, 0)
	ch.GetHeading().SetFont(fontRegular)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	sc := c.NewSubchapter(ch, "Issuer details")
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	issuerTable := creator.NewTable(2)
	issuerTable.SetMargins(0, 0, 30, 0)

	pColor := creator.ColorRGBFrom8bit(72, 86, 95)
	bgColor := creator.ColorRGBFrom8bit(56, 68, 67)

	p := creator.NewParagraph("Issuer")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell := issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = creator.NewParagraph("UniDoc")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	p = creator.NewParagraph("Address")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = creator.NewParagraph("Klapparstig 16, 101 Reykjavik, Iceland")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	p = creator.NewParagraph("Email")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBackgroundColor(bgColor)
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	p = creator.NewParagraph("sales@unidoc.io")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	p = creator.NewParagraph("Web")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = creator.NewParagraph("unidoc.io")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	p = creator.NewParagraph("Author")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = creator.NewParagraph("UniDoc report generator")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderStyleBox, 1)
	cell.SetContent(p)

	sc.Add(issuerTable)

	// 1.2 - Document history
	sc = c.NewSubchapter(ch, "Document History")
	sc.SetMargins(0, 0, 5, 0)
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(pColor)

	histTable := creator.NewTable(3)
	histTable.SetMargins(0, 0, 30, 50)

	histCols := []string{"Date Issued", "UniDoc Version", "Type/Change"}
	for _, histCol := range histCols {
		p = creator.NewParagraph(histCol)
		p.SetFont(fontBold)
		p.SetFontSize(10)
		p.SetColor(creator.ColorWhite)
		cell = histTable.NewCell()
		cell.SetBackgroundColor(bgColor)
		cell.SetBorder(creator.CellBorderStyleBox, 1)
		cell.SetHorizontalAlignment(creator.CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(creator.CellVerticalAlignmentMiddle)
		cell.SetContent(p)
	}

	dateStr := unicommon.ReleasedAt.Format("1 Jan, 2006 15:04")

	histVals := []string{dateStr, unicommon.Version, "First issue"}
	for _, histVal := range histVals {
		p = creator.NewParagraph(histVal)
		p.SetFont(fontRegular)
		p.SetFontSize(10)
		p.SetColor(pColor)
		cell = histTable.NewCell()
		cell.SetBorder(creator.CellBorderStyleBox, 1)
		cell.SetHorizontalAlignment(creator.CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(creator.CellVerticalAlignmentMiddle)
		cell.SetContent(p)
	}

	sc.Add(histTable)

	err := c.Draw(ch)
	if err != nil {
		panic(err)
	}
}

// Chapter giving an overview of features.
// TODO: Add code snippets and show more styles and options.
func DoFeatureOverview(c *creator.Creator, fontRegular *model.PdfFont, fontBold *model.PdfFont) {
	// Ensure that the chapter starts on a new page.
	c.NewPage()

	ch := c.NewChapter("Feature overview")

	chapterFont := fontRegular
	chapterFontColor := creator.ColorRGBFrom8bit(72, 86, 95)
	chapterFontSize := 18.0

	normalFont := fontRegular
	normalFontColor := creator.ColorRGBFrom8bit(72, 86, 95)
	normalFontSize := 10.0

	bgColor := creator.ColorRGBFrom8bit(56, 68, 67)

	ch.GetHeading().SetFont(chapterFont)
	ch.GetHeading().SetFontSize(chapterFontSize)
	ch.GetHeading().SetColor(chapterFontColor)

	p := creator.NewParagraph("This chapter demonstrates a few of the features of UniDoc that can be used for report generation.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	ch.Add(p)

	// Paragraphs.
	sc := c.NewSubchapter(ch, "Paragraphs")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("Paragraphs are used to represent text, as little as a single character, a word or " +
		"multiple words forming multiple sentences. UniDoc handles automatically wrapping those across lines and pages, making " +
		"it relatively easy to work with. They can also be left, center, right aligned or justified as illustrated below:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	// Example paragraphs:
	loremTxt := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum."
	alignments := []creator.TextAlignment{creator.TextAlignmentLeft, creator.TextAlignmentCenter,
		creator.TextAlignmentRight, creator.TextAlignmentJustify}
	for j := 0; j < 4; j++ {
		p = creator.NewParagraph(loremTxt)
		p.SetFont(normalFont)
		p.SetFontSize(normalFontSize)
		p.SetColor(normalFontColor)
		p.SetMargins(20, 0, 10, 10)
		p.SetTextAlignment(alignments[j%4])

		sc.Add(p)
	}

	sc = c.NewSubchapter(ch, "Tables")
	// Mock table: Priority table.
	priTable := creator.NewTable(2)
	priTable.SetMargins(40, 40, 10, 0)
	// Column headers:
	tableCols := []string{"Priority", "Items fulfilled / available"}
	for _, tableCol := range tableCols {
		p = creator.NewParagraph(tableCol)
		p.SetFont(fontBold)
		p.SetFontSize(10)
		p.SetColor(creator.ColorWhite)
		cell := priTable.NewCell()
		cell.SetBackgroundColor(bgColor)
		cell.SetBorder(creator.CellBorderStyleBox, 1)
		cell.SetContent(p)
	}
	items := [][]string{
		[]string{"High", "52/80"},
		[]string{"Medium", "32/100"},
		[]string{"Low", "10/90"},
	}
	for _, lineItems := range items {
		for _, item := range lineItems {
			p = creator.NewParagraph(item)
			p.SetFont(fontBold)
			p.SetFontSize(10)
			p.SetColor(creator.ColorWhite)
			cell := priTable.NewCell()
			cell.SetBackgroundColor(bgColor)
			cell.SetBorder(creator.CellBorderStyleBox, 1)
			cell.SetContent(p)
		}
	}
	sc.Add(priTable)

	sc = c.NewSubchapter(ch, "Images")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("Images can be loaded from multiple file formats, example from a PNG image:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 5)
	sc.Add(p)

	// Show logo.
	img, err := creator.NewImageFromFile("./unidoc-logo.png")
	if err != nil {
		panic(err)
	}
	img.ScaleToHeight(50)
	sc.Add(img)

	sc = c.NewSubchapter(ch, "QR Codes / Barcodes")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("Example of a QR code generated with package github.com/boombuler/barcode:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 5)
	sc.Add(p)

	qrCode, _ := makeQrCodeImage("HELLO", 40, 5)
	img, err = creator.NewImageFromGoImage(qrCode)
	if err != nil {
		panic(err)
	}
	img.SetWidth(40)
	img.SetHeight(40)
	sc.Add(img)

	sc = c.NewSubchapter(ch, "Graphing / Charts")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("Graphs can be generated via packages such as github.com/wcharczuk/go-chart as illustrated " +
		"in the following plot:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	graph := chart.PieChart{
		Width:  200,
		Height: 200,
		Values: []chart.Value{
			{Value: 70, Label: "Compliant"},
			{Value: 30, Label: "Non-Compliant"},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}
	img, err = creator.NewImageFromData(buffer.Bytes())
	if err != nil {
		panic(err)
	}
	img.SetMargins(0, 0, 10, 0)
	sc.Add(img)

	sc = c.NewSubchapter(ch, "Headers and footers")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("Convenience functions are provided to generate headers and footers, see: " +
		"https://godoc.org/github.com/unidoc/unidoc/pdf/creator#Creator.DrawHeader and " +
		"https://godoc.org/github.com/unidoc/unidoc/pdf/creator#Creator.DrawFooter " +
		"They both set a function that accepts a block which the header/footer is drawn on for each page. " +
		"More information is provided in the arguments, allowing to skip header/footer on specific pages and " +
		"showing page number and count.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	sc = c.NewSubchapter(ch, "Table of contents generation")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = creator.NewParagraph("A convenience function is provided to generate table of contents " +
		"as can be seen on https://godoc.org/github.com/unidoc/unidoc/pdf/creator#Creator.CreateTableOfContents and " +
		"in our example code on unidoc.io.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	c.Draw(ch)
}

// Helper function to make the QR code image with a specified oversampling factor.
// The oversampling specifies how many pixels/point. Standard PDF resolution is 72 points/inch.
func makeQrCodeImage(text string, width float64, oversampling int) (goimage.Image, error) {
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}

	pixelWidth := oversampling * int(math.Ceil(width))
	qrCode, err = barcode.Scale(qrCode, pixelWidth, pixelWidth)
	if err != nil {
		return nil, err
	}

	return qrCode, nil
}
