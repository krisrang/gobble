import Ember from 'ember';
import config from './config/environment';

var Router = Ember.Router.extend({
  location: config.locationType
});

Router.map(function(){
  this.route('index', { path: '/' });
  this.route('wwf', { path: '/wwf/:tiles' });
  this.route('scrabble', { path: '/scrabble/:tiles' });
});

export default Router;
