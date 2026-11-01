package mapgen

// styles lists the age rows, oldest first: the picture at the close zoom,
// which every age has.
var styles = composeAges(ageRows)

// farStyles lists the picture at the far zoom for the ages that change with
// it: the star map of the space ages.
var farStyles = composeAges(spaceRows)

// firstFarAge is the oldest age that draws the star map at the far zoom.
const firstFarAge = "space_age"
