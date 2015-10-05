export default Ember.Helper.helper(function(params, namedArgs) {
  let word = params[0],
      tiles = Ember.Handlebars.Utils.escapeExpression(namedArgs.tiles),
      result = "";

  tiles = tiles.replace(/\W/g, '').toLowerCase();
  for (var i = 0; i < word.length; i++) {
    var letter = word.charAt(i);

    if (tiles.indexOf(letter.toLowerCase()) === -1) {
      result += `<span class="wildcard">${letter}</span>`;
    } else {
      result += `<span class="letter">${letter}</span>`;;
    }
  }

  return Ember.String.htmlSafe(`${result}`);
});
