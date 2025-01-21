let history = []; // Store daily feedback for learning

// Function to calculate heating time based on learning from previous feedback
function calculateHeatingTime(day) {
    // Base heating time calculation
    let baseHeatingTime = day.showerDuration + 10; // Shower duration plus buffer
    let temperatureFactor = (30 - day.averageTemperature) / 2; // Temperature adjustment
    let heatingTime = baseHeatingTime + temperatureFactor;

    // Learn from previous feedback
    if (history.length > 0) {
        // Get average adjustments from last 5 days or all available history
        const recentHistory = history.slice(-5);
        let totalAdjustment = 0;
        let contributingEntries = 0;

        recentHistory.forEach(entry => {
            // Only adjust if the previous result wasn't perfect (satisfaction !== 5)
            if (entry.satisfaction !== 5) {
                contributingEntries++;
                // Calculate adjustment based on satisfaction (1-10 scale)
                if (entry.satisfaction < 5) {
                    // Too cold - increase heating time
                    totalAdjustment += (5 - entry.satisfaction) * 2;
                } else if (entry.satisfaction > 5) {
                    // Too hot - decrease heating time
                    totalAdjustment -= (entry.satisfaction - 5) * 2;
                }
                
                // Consider temperature similarity
                const tempDiff = Math.abs(entry.averageTemperature - day.averageTemperature);
                if (tempDiff < 5) { // If temperatures are similar
                    totalAdjustment *= 1.5; // Give more weight to similar conditions
                }
            }
        });

        // Apply the learned adjustment
        heatingTime += contributingEntries > 0 ? totalAdjustment / contributingEntries : 0;
    }

    return Math.max(20, Math.min(80, heatingTime)); // Limit between 20-80 minutes
}

// Function to record feedback and improve future suggestions
function provideFeedback(day, satisfaction) {
    if (satisfaction < 1 || satisfaction > 10) {
        console.log("Please provide feedback between 1 (too cold) and 10 (too hot)");
        return;
    }

    // Store the feedback and conditions
    day.satisfaction = satisfaction;
    day.actualHeatingTime = calculateHeatingTime(day);
    history.push(day);

    console.log(`Day ${history.length}: Conditions: Temp=${day.averageTemperature}°C, Duration=${day.showerDuration}min`);
    console.log(`Suggested heating time: ${day.actualHeatingTime.toFixed(2)} minutes`);
    console.log(`Feedback recorded: ${satisfaction}/10 (${satisfaction < 5 ? 'too cold' : satisfaction > 5 ? 'too hot' : 'perfect'})\n`);
}


// Example usage
let day1 = {
    averageTemperature: 20,
    showerDuration: 30,
};

let heatingTime = calculateHeatingTime(day1);
provideFeedback(day1, 3); // Too cold

let day2 = {
    averageTemperature: 20,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day2);
provideFeedback(day2, 6); // Just right

let day3 = {
    averageTemperature: 20,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day3);
provideFeedback(day3, 5); // Too hot

let day4 = {
    averageTemperature: 25,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day4);
provideFeedback(day4, 6); // Slightly too hot

let day5 = {
    averageTemperature: 25,
    showerDuration: 30,
};
heatingTime = calculateHeatingTime(day5);
provideFeedback(day5, 5); // Just right
