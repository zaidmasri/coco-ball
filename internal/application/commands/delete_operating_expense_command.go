package commands

import "uuid"

// DeleteOperatingExpense is the command to delete an Operating Expenses
// wizard item.
type DeleteOperatingExpense struct {
	ItemID uuid.UUID
}

// DeleteOperatingExpenseResult reports whether the deletion succeeded.
type DeleteOperatingExpenseResult struct {
	Success bool
}
