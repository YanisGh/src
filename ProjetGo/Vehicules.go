package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Struct to hold data from the API response
type Vehicle struct {
	Make      string `json:"make"`
	Model     string `json:"model"`
	Year      string `json:"year"`
	Cylinders int    `json:"cylinders"`
}

// Fetch vehicles based on dynamic parameters
func fetchVehicles(make, model, sort string, year, cylinders, resultNb int) ([]Vehicle, error) {
	// Base API URL
	apiURL := "https://public.opendatasoft.com/api/records/1.0/search/?dataset=all-vehicles-model&q="

	fmt.Print("ANNEE : ", year, "CYLINDER : ", cylinders)
	if resultNb != 0 {
		apiURL += "&rows=" + fmt.Sprintf("%d", resultNb)
	} else {
		// 10 cars by default
		apiURL += "&rows=10"
	}

	if sort != "" {
		apiURL += "&sort=" + sort
	}

	// Only ask for the model if the make is not empty

	if make != "" {
		apiURL += "&refine.make=" + make
	}

	if model != "" {
		apiURL += "&refine.model=" + model
	}

	if year != 0 {
		apiURL += "&refine.year=" + fmt.Sprintf("%d", year)
	}
	if cylinders != 0 {
		apiURL += "&refine.cylinders=" + fmt.Sprintf("%d", cylinders)
	}

	fmt.Println("URL fini " + apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from external API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned an error: %v", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract records
	var vehicles []Vehicle
	records, ok := result["records"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure from API")
	}

	for _, record := range records {
		data, ok := record.(map[string]interface{})["fields"].(map[string]interface{})
		if !ok {
			continue
		}
		make, _ := data["make"].(string)
		model, _ := data["model"].(string)
		year, _ := data["year"].(string)
		cylinders, _ := data["cylinders"].(float64) // Assert to float64 first
		cylindersInt := int(cylinders)              // Convert to int
		//fmt.Println("CYLINDERS: ", cylindersInt)

		vehicles = append(vehicles, Vehicle{
			Make:      make,
			Model:     model,
			Year:      year,
			Cylinders: cylindersInt, // Now it's an int
		})
	}

	return vehicles, nil
}

func saveVehiclesToJSON(vehicles []Vehicle, filename string) error {
	// Convert the vehicles slice to JSON
	jsonData, err := json.MarshalIndent(vehicles, "", "  ") // Pretty print JSON
	if err != nil {
		return fmt.Errorf("failed to marshal vehicles to JSON: %v", err)
	}

	// Create or overwrite the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %v", err)
	}
	defer file.Close()

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	fmt.Printf("Vehicles data saved to %s\n", filename)
	return nil
}

func saveVehiclesToCSV(vehicles []Vehicle, filename string) error {
	// Create or overwrite the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	header := []string{"Make", "Model", "Year", "Cylinders"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header to CSV: %v", err)
	}

	// Write each vehicle as a row
	for _, vehicle := range vehicles {
		row := []string{
			vehicle.Make,
			vehicle.Model,
			vehicle.Year,
			strconv.Itoa(vehicle.Cylinders), // Convert int to string for CSV
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row to CSV: %v", err)
		}
	}

	fmt.Printf("Vehicles data saved to %s\n", filename)
	return nil
}

func main() {
	var make, model, sort string
	var cylinders, year, resultNb int

	// Ask the user for input
	fmt.Println("Enter the car maker (press Enter to skip):")
	fmt.Scanln(&make)

	if make != "" {
		fmt.Println("Enter the model (press Enter to skip):")
		fmt.Scanln(&model)
	}

	for {
		year = 0
		fmt.Println("Enter the year (press Enter to skip):")
		fmt.Scanln(&year)
		//fmt.Println("year :", year)

		if year == 0 || year >= 1980 && year <= 2025 {
			break
		}
		fmt.Println("Please select a year between 1980 and 2025.")
	}

	for {
		cylinders = 0
		fmt.Println("Enter the number of cylinders (press Enter to skip):")
		fmt.Scanln(&cylinders)
		//fmt.Print("nb cylindres", cylinders)

		if (cylinders == 0 || cylinders >= 3 && cylinders <= 6) || cylinders == 8 || cylinders == 10 || cylinders == 12 || cylinders == 16 {
			break
		}
		fmt.Println("Invalid option.")
	}

	for {
		resultNb = 0
		fmt.Println("Enter the maximum number of results you want (press Enter to skip):")
		fmt.Scanln(&resultNb)
		//fmt.Println("nb result", resultNb)

		if resultNb >= 0 && resultNb <= 50 {
			break
		}
		fmt.Println("Invalid option.")
	}

	for {
		sort = ""
		fmt.Println("Do you want to sort the results? (Sort by: make, model, year, cylinders) (press Enter to skip):")
		fmt.Scanln(&sort)

		if sort == "" || sort == "make" || sort == "model" || sort == "year" || sort == "cylinders" {
			break
		}
		fmt.Println("Invalid option.")
	}
	// Fetch vehicles based on user input
	vehicles, err := fetchVehicles(make, model, sort, year, cylinders, resultNb)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Display the results
	if len(vehicles) == 0 {
		fmt.Println("No vehicles found for the given criteria.")
	} else {
		fmt.Println("Vehicles found:")
		for _, v := range vehicles {
			fmt.Printf("Make: %s, Model: %s, Year: %s, Cylinders: %d\n", v.Make, v.Model, v.Year, v.Cylinders)
		}

		// Save vehicles to a JSON file
		err := saveVehiclesToJSON(vehicles, "vehicles.json")
		if err != nil {
			fmt.Println("Error saving to JSON:", err)
		}

		// Save vehicles to a CSV file
		err = saveVehiclesToCSV(vehicles, "vehicles.csv") // Use `=` to reassign
		if err != nil {
			fmt.Println("Error saving to CSV:", err)
		}
	}
}
