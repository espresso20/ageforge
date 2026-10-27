package mapgen

// styles lists the age rows, oldest first: the ground and city ages, then the
// space rows.
var styles = append(composeAges(ageRows), space)
