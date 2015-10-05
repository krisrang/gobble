import Ember from 'ember';

export default Ember.Controller.extend({
  disableSearch: Ember.computed.empty('searchString'),
  isScrabble: Ember.computed.equal('mode', 'scrabble'),
  isWWF: Ember.computed.equal('mode', 'wwf'),

  actions: {
    search() {
      var search = this.get('searchString').replace(/[^A-Za-z+]/g, '');
      this.transitionToRoute('/' + this.get('mode') + '/' + search);
    },

    setMode(mode) {
      this.set('mode', mode);
      this.send('search');
    }
  }
});
