// isFunction returns true if the provided argument is a function
// credits: https://stackoverflow.com/questions/5999998/check-if-a-variable-is-of-function-type
export function isFunction(functionToCheck) {
  return functionToCheck && {}.toString.call(functionToCheck) === '[object Function]';
}

export function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

const SI_SYMBOL = ['', 'k', 'M', 'G', 'T', 'P', 'E'];
// credits: https://stackoverflow.com/questions/9461621/format-a-number-as-2-5k-if-a-thousand-or-more-otherwise-900
export function abbreviateNumber(number) {
  // what tier? (determines SI symbol)
  // eslint-disable-next-line no-bitwise
  const tier = Math.log10(Math.abs(number)) / 3 | 0;

  // if zero, we don't need a suffix
  if (tier === 0) return number;

  // get suffix and determine scale
  const suffix = SI_SYMBOL[tier];
  const scale = 10 ** (tier * 3);

  // scale the number
  const scaled = number / scale;

  // format number and add suffix
  return scaled.toFixed(1) + suffix;
}
