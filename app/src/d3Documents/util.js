// isFunction returns true if the provided argument is a function
// credits: https://stackoverflow.com/questions/5999998/check-if-a-variable-is-of-function-type
export function isFunction(functionToCheck) {
  return functionToCheck && {}.toString.call(functionToCheck) === '[object Function]';
}

export function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
