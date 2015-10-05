export default Ember.Helper.helper(function(params) {
  let word = params[0],
      length = 0;

  length = word.length;
  return `${length}`;
});
