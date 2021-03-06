import Ember from 'ember';

export default Ember.Route.extend({
  ajax: Ember.inject.service(),

  model(params) {
    return this.get('ajax').request('/api/scrabble/' + params.tiles).then(function(result){
      return {tiles: params.tiles, result: result};
    });
  },

  setupController(controller, model) {
    var appController = this.controllerFor('application');
    appController.set('mode', 'scrabble');
    appController.set('searchString', model.tiles);

    controller.set('model', model);
  }
})
