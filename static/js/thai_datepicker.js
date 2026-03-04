document.addEventListener("DOMContentLoaded", function () {
    flatpickr(".datepicker-th", {
        locale: "th",
        dateFormat: "Y-m-d", // Store as YYYY-MM-DD (CE)
        disableMobile: true, // Always use flatpickr UI, never native mobile picker
        allowInput: false,
        altInput: true,      // Show user-friendly format
        altFormat: "d-m-Y",   // Display as DD-MM-YYYY (converted to BE by formatDate)
        onReady: function (selectedDates, dateStr, instance) {
            const yearInput = instance.currentYearElement;
            if (yearInput) {
                // Adjust year display in calendar header to BE
                yearInput.value = parseInt(yearInput.value) + 543;
            }
        },
        onYearChange: function (selectedDates, dateStr, instance) {
            // Adjust year display in calendar header to BE when year changes
            let year = instance.currentYear;

            // Check if user likely entered a BE year (e.g., 2568)
            if (year > 2200) {
                year = year - 543;
                instance.currentYear = year; // Update internal year to AD
                instance.redraw(); // Refresh calendar to show correct AD year days
            }

            const yearInput = instance.currentYearElement;
            if (yearInput) {
                setTimeout(() => {
                    yearInput.value = year + 543; // Always display as BE
                }, 10);
            }
        },
        onMonthChange: function (selectedDates, dateStr, instance) {
            // Adjust year display in calendar header to BE when month changes (which might change year)
            const yearInput = instance.currentYearElement;
            if (yearInput) {
                setTimeout(() => {
                    yearInput.value = parseInt(instance.currentYear) + 543;
                }, 10);
            }
        },
        formatDate: (date, format, locale) => {
            // Custom formatter for Thai Year (BE)
            if (format === "d-m-Y") {
                const year = date.getFullYear() + 543;
                const day = String(date.getDate()).padStart(2, '0');
                const month = String(date.getMonth() + 1).padStart(2, '0');
                return `${day}-${month}-${year}`;
            }
            return flatpickr.formatDate(date, format);
        }
    });
});
