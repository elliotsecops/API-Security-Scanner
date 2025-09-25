start: 
	@bash ./start.sh

start-nogui:
	@bash ./start.sh --no-gui

build-gui:
	npm --prefix ./gui install && npm --prefix ./gui run build
