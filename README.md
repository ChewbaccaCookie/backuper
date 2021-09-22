# backuper


 
CRONTAB:

0 * * * * /home/pi/scripts/backuper hour >> /home/pi/scripts/logs/backup-hour.log 2>&1

0 1 * * * /home/pi/scripts/backuper day >> /home/pi/scripts/logs/backup-day.log 2>&1

0 2 * * 0 /home/pi/scripts/backuper week >> /home/pi/scripts/logs/backup-week.log 2>&1

0 3 1 * * /home/pi/scripts/backuper month >> /home/pi/scripts/logs/backup-month.log 2>&1

0 4 1 1 * /home/pi/scripts/backuper year >> /home/pi/scripts/logs/backup-year.log 2>&1
