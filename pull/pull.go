package pull

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gocql/gocql"
	"github.com/jochenboesmans/forward-financial-statements/tickers"
)

type IncomeStatementTimeSeries []IncomeStatement
type IncomeStatement struct {
	Revenue              float64 `json:"revenue"`
	NetIncome            float64 `json:"netIncome"`
	GrossProfitRatio     float64 `json:"grossProfitRatio"`
	EbitdaRatio          float64 `json:"ebitdaratio"`
	OperatingIncomeRatio float64 `json:"operatingIncomeRatio"`
	IncomeBeforeTaxRatio float64 `json:"incomeBeforeTaxRatio"`
	NetIncomeRatio       float64 `json:"NetIncomeRatio"`
}

func getIncomeStatements(ticker string, apiKey string) IncomeStatementTimeSeries {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/income-statement/%s?period=quarter&limit=400&apikey=%s", ticker, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}

	r := IncomeStatementTimeSeries{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}
	r.reverse()

	return r
}

func (ists IncomeStatementTimeSeries) reverse() {
	r := IncomeStatementTimeSeries{}
	for i := range ists {
		r = append(r, ists[len(ists)-1-i])
	}

	for i := range r {
		ists[i] = r[i]
	}
}

func (is IncomeStatement) access(property string) (float64, error) {
	switch property {
	case "Revenue":
		return is.Revenue, nil
	case "NetIncome":
		return is.NetIncome, nil
	case "GrossProfitRatio":
		return is.GrossProfitRatio, nil
	case "OperatingIncomeRatio":
		return is.OperatingIncomeRatio, nil
	case "IncomeBeforeTaxRatio":
		return is.IncomeBeforeTaxRatio, nil
	case "NetIncomeRatio":
		return is.NetIncomeRatio, nil
	case "EbitdaRatio":
		return is.EbitdaRatio, nil
	default:
		return 0.0, errors.New("invalid property name")
	}
}

func (ists IncomeStatementTimeSeries) Select(property string) []float64 {
	r := []float64{}
	for _, is := range ists {
		if propertyValue, err := is.access(property); err == nil {
			r = append(r, propertyValue)
		}
	}
	return r
}

func Pull(dbSession *gocql.Session) {
	apiKey := os.Getenv("API_KEY")
	// TODO: error handling
	ts, _ := tickers.ReadTickersFromDb(dbSession)

	incomeStatements := map[string]IncomeStatementTimeSeries{}
	for _, t := range ts {
		incomeStatements[t] = getIncomeStatements(t, apiKey)
	}

	incomeStatementsJSON, err := json.Marshal(incomeStatements)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile("financial-statements.json", incomeStatementsJSON, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("all done")
}
