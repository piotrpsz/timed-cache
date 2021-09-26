# timed-cache
A cache that stores items for a specified period of time. Older items are removed.<br>
Expired items are removed before almost every function of the cache object is executed.<br>
Additionally, the user has the option of manually clearing the cache of expired items.<br>
If you need periodic cleanup without even performing an operation,<br>
implement a ticker that calls PurgeExpired in code that uses the cache.<br>
