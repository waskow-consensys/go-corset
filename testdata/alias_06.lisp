(defcolumns COUNTER)
(defalias CT3 CT2)
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () CT3)