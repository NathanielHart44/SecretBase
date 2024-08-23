	1.	If the .env.example file does not exist, it is created with all the keys from the .env file, each with an empty value.
	2.	If a key from the .env file is already present in the .env.example file, it should not be overwritten.
	3.	If a key from the .env file is not present in the .env.example file, it should be added with the value ''.
    4.  If an old key is present in the .env.example file but not in the .env file, it should be removed from the .env.example file.