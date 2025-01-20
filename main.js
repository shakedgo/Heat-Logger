let history = []; // To store daily feedback for learning

// Function to calculate heating time based on the current formula
function calculateHeatingTime(day) {
    // Base heating time: Shower duration + buffer + temperature adjustment
    let baseHeatingTime = day.showerDuration + 10; // Default: Shower duration + buffer
    let temperatureFactor = (30 - day.averageTemperature) / 2; // Lower temp → more heating needed
    let heatingTime = baseHeatingTime + temperatureFactor;

    // Learn from history: Update formula based on the margin of error from previous feedback
    if (history.length > 0) {
        const lastEntry = history[history.length - 1]; // Most recent entry
        const feedback = lastEntry.satisfaction;

        // Calculate the margin of error (suggested time - actual time)
        const marginOfError = Math.abs(heatingTime - lastEntry.actualHeatingTime);

        // Adjust the formula based on the margin of error
        if (marginOfError > 5) {
            // If the error is too large, refine the temperature factor and duration buffer
            const adjustmentFactor = marginOfError / 10;
            if (feedback < 5) {
                // If the feedback is too cold, increase the temperature weight
                temperatureFactor += adjustmentFactor;
            } else if (feedback > 5) {
                // If the feedback is too hot, decrease the temperature weight
                temperatureFactor -= adjustmentFactor;
            }
        }
    }

    return Math.max(heatingTime, baseHeatingTime); // Ensure minimum heating time
}

// Function to provide feedback and adjust the formula
function provideFeedback(day, satisfaction) {
    day.satisfaction = satisfaction;

    // Calculate the actual heating time based on satisfaction
    // 1 means too cold, 10 means too hot, 5 means just right
    let actualHeatingTime = day.showerDuration + 10 + (30 - day.averageTemperature) / 2;
    if (satisfaction < 5) {
        // If too cold, actual time should be higher
        actualHeatingTime += (5 - satisfaction) * 2;
    } else if (satisfaction > 5) {
        // If too hot, actual time should be lower
        actualHeatingTime -= (satisfaction - 5) * 2;
    }

    // Store feedback and actual heating time for future learning
    day.actualHeatingTime = actualHeatingTime;
    history.push(day);

    console.log(`Feedback received - Satisfaction: ${satisfaction}. Actual Heating Time: ${actualHeatingTime}`);
}



// Example usage
let day1 = {
    averageTemperature: 20,
    showerDuration: 30,
};

let heatingTime = calculateHeatingTime(day1);
console.log(`Day 1: Average temperature: ${day1.averageTemperature}°C, Shower duration: ${day1.showerDuration} minutes`);
console.log(`Suggested heating time: ${heatingTime.toFixed(2)} minutes`);
provideFeedback(day1, 3);

let day2 = {
    averageTemperature: 20,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day2);
console.log(`Day 2: Average temperature: ${day2.averageTemperature}°C, Shower duration: ${day2.showerDuration} minutes`);
console.log(`Suggested heating time: ${heatingTime.toFixed(2)} minutes`);
provideFeedback(day2, 5); // Example: 4 means just right

let day3 = {
    averageTemperature: 20,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day3);
console.log(`Day 3: Average temperature: ${day3.averageTemperature}°C, Shower duration: ${day3.showerDuration} minutes`);
console.log(`Suggested heating time: ${heatingTime.toFixed(2)} minutes`);
provideFeedback(day3, 5); // Example: 4 means just right

let day4 = {
    averageTemperature: 30,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day3);
console.log(`Day 4: Average temperature: ${day3.averageTemperature}°C, Shower duration: ${day3.showerDuration} minutes`);
console.log(`Suggested heating time: ${heatingTime.toFixed(2)} minutes`);
provideFeedback(day3, 8); // Example: 4 means just right
