# WotWizard

WotWizard builds, from the blockchain and sandbox of Duniter, a prediction of the future entries of candidates into the Duniter Web of Trust (WOT). It uses a simulation of the Duniter mechanism. When several possibilities may happen, each one is listed with its probabilities. The published lists are automatically updated every five minutes, and any change is signalled visually.

WotWizard is divided into:
    - a server (WWServer.exe) which produces two data files (WWByDates.json & WWMeta.json)
    - a gui interface (WotWizard.exe) which reads the data from WWServer.exe and displays them

You can find the executables here:
	https://github.com/duniter/WotWizard/releases


This program needs that a Duniter node runs on the same computer.
This program runs natively on Windows. If your computer runs on Linux, install "wine" first, and, in a terminal, run "winecfg" and configure the drives to be sure that the Duniter database (see below) can be reached.

You can run WotWizard as a server only, if you want to display yourself its results (e.g. in a web page). In this case, put the file "WWServer.exe" into an empty directory and put the Windows dll "sqlite3.dll" into the same directory. Then, you have to run "WWServer.exe" as often as you want with the line command:
	
	$wine WWServer.exe
	
or directly from within your application. The dll can be found at the address:
	
	https://www.sqlite.org/2015/sqlite-dll-win32-x86-3081002.zip

At the end of every run, WWServer writes (or updates) two files, in json syntax: WWByDates.json, which lists the entries of the newcomers, sorted and grouped by dates, and WWMete.json, which gives some metadata on the computing and the original data.

You have to tell WWServer two things:
	- the path where it can find the Duniter database.
		By default: 
			D:\.config\duniter\duniter_default\duniter.db
		Change the default value by creating (or editing) the file Duniter/Rsrc/Init.txt; warning: the path must be surrounded by two double quotes (ex: "D:\.config\duniter\duniter_default\duniter.db")
	- the largest memory size in bytes (approximatively) it is allowed to allocate.
		By default: 800000000
		Change the default value by creating (or editing) the file Duniter/Rsrc/WW_Max_Stack.txt.

You can change the language of the output with these line commands options:
	
	$wine WWServer.exe -lang fr	for a French output, or
	$wine WWServer.exe -lang en	for an English output.

On the first run of WWServer, a part of Duniter data is copied into a new database (Duniter/DBase.data), to accelerate their future use. This operation may take a rather long time,.

If you want to display the updated results of WWServer continuously, put the file "WotWizard.exe" into the same directory as "WWServer.exe" and run WotWizard with the line command:
	
	$wine WotWizard.exe

You can:
	- choose the way the list is displayed (by names or by dates, or metadata)
	- manually update the list (it's automatically updated every five minutes)

When the list has changed, two asterisks appear, one on each side of the title, and a new button "Check" is created. Click on the button to make the marks disappear. You can then compare the new and old lists by clicking on the button "Compare", or by using the menu item "Edit -> Compare Texts".

