const { dest, series, src, watch } = require('gulp');

const clean = require('gulp-clean');
const concat = require('gulp-concat');
const gulpif = require('gulp-if');
const jshint = require('gulp-jshint');
const minify = require('gulp-babel-minify');
const purgeSourcemaps = require('gulp-purge-sourcemaps');
const removeEmptyLines = require('gulp-remove-empty-lines');
const sourcemaps = require('gulp-sourcemaps');
const sass = require('gulp-sass')(require('sass'));

function cleanup() {
	return src('resources/static/web.css', {"allowEmpty": true})
		.pipe(clean());
}

function css() {
	return src('resources/web.scss')
		.pipe(sass({outputStyle: 'compressed'})
			.on('error', sass.logError))
		.pipe(dest('resources/static/'))
}

function lint() {
	return src('resources/web.js')
		.pipe(jshint())
		.pipe(jshint.reporter('jshint-stylish'))
		.pipe(jshint.reporter('fail'));
}

function js() {
	return src('node_modules/bootstrap/dist/js/bootstrap.bundle.min.js')
		.pipe(src('resources/web.js'))
		.pipe(gulpif('!**/*.min.js', minify()))
		.pipe(sourcemaps.init({loadMaps: true}))
		.pipe(purgeSourcemaps())
		.pipe(concat('web.js'))
		.pipe(removeEmptyLines())
		.pipe(dest('resources/static/'));
}

exports.clean = cleanup;
exports.css = css;
exports.js = series(lint, js);
exports.lint = lint;

exports.watch = function() {
	watch('resources/web.scss', css);
	watch('resources/web.js', series(lint, js));
};

exports.default = series(cleanup, css, lint, js)
