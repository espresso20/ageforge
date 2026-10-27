package mapgen

// styles lists the age rows, oldest first: the ground and city ages, then the
// space ages.
var styles = composeAges(append(append([]AgeRow(nil), ageRows...), spaceRows...))
