--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: account_exists(character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT EXISTS(SELECT 1 FROM i9c_user WHERE email_or_username = ANY(ARRAY[email, username])) INTO exist;
END;
$$;


ALTER FUNCTION public.account_exists(email_or_username character varying, OUT exist boolean) OWNER TO postgres;

--
-- Name: add_users_to_group(integer, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.add_users_to_group(in_group_chat_id integer, in_admin character varying[], in_users character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  nuser varchar[]; -- ['{id}', '{username}']
  nusers_usernames varchar[];
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  FOREACH nuser SLICE 1 IN ARRAY in_users
  LOOP
    -- create user_group_chat for in_users
    INSERT INTO user_group_chat (user_id, group_chat_id)
    VALUES (nuser[1]::int, in_group_chat_id);
    
    -- create group_chat_membership for in_users
	INSERT INTO group_chat_membership (group_chat_id, member_id)
	VALUES (in_group_chat_id, nuser[1]::int);
    
    -- extract in_users' username
	nusers_usernames := array_append(nusers_usernames, nuser[2]);
  END LOOP;
  
  -- increment members_count in group_chat
  UPDATE group_chat 
  SET members_count = members_count + array_length(in_users, 1)
  WHERE id = in_group_chat_id;
  
  -- create group_chat_activity_log for users added
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'users_added', json_build_object('added_by', in_admin[2], 'new_users', nusers_usernames))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.add_users_to_group(in_group_chat_id integer, in_admin character varying[], in_users character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: change_group_description(integer, character varying[], character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.change_group_description(in_group_chat_id integer, in_admin character varying[], in_new_description character varying, OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set group_chat's description to new description
  UPDATE group_chat 
  SET description = in_new_description
  WHERE id = in_group_chat_id;
  
  -- create group_chat_activity_log for group description change
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'admin_changed_group_description', json_build_object('admin_name', in_admin[2], 'new_group_description', in_new_description))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.change_group_description(in_group_chat_id integer, in_admin character varying[], in_new_description character varying, OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: change_group_name(integer, character varying[], character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.change_group_name(in_group_chat_id integer, in_admin character varying[], in_new_name character varying, OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set group_chat's name to new name
  UPDATE group_chat 
  SET "name" = in_new_name
  WHERE id = in_group_chat_id;
  
  -- create group_chat_activity_log for group name change
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'admin_changed_group_name', json_build_object('admin_name', in_admin[2], 'new_group_name', in_new_name))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.change_group_name(in_group_chat_id integer, in_admin character varying[], in_new_name character varying, OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: change_group_picture(integer, character varying[], character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.change_group_picture(in_group_chat_id integer, in_admin character varying[], in_new_picture character varying, OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set group_chat's name to new picture
  UPDATE group_chat 
  SET picture = in_new_picture
  WHERE id = in_group_chat_id;
  
  -- create group_chat_activity_log for group picture change
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'admin_changed_group_picture', json_build_object('admin_name', in_admin[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.change_group_picture(in_group_chat_id integer, in_admin character varying[], in_new_picture character varying, OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: edit_user(integer, character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.edit_user(in_user_id integer, in_col_updates character varying[]) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $_$
DECLARE
  col_name_val varchar[];
  update_sets varchar := '';
BEGIN
  FOREACH col_name_val SLICE 1 IN ARRAY in_col_updates 
  LOOP
    IF col_name_val[1] NOT IN ('username', 'password', 'email', 'profile_picture', 'location') THEN
	  RAISE EXCEPTION '"%" is either an invalid or a non-editable column', col_name_val[1] 
	  USING HINT = 'Validate column name or set column from the proper routine';
	END IF;
    update_sets := update_sets || col_name_val[1] || ' = ''' || col_name_val[2] || ''', ';
  END LOOP;
  
  update_sets := LEFT(update_sets, LENGTH(update_sets) - 2);
  
  RETURN QUERY EXECUTE 'UPDATE i9c_user SET ' || update_sets || ' WHERE id = $1 RETURNING i9c_user.id, i9c_user.username, i9c_user.profile_picture, i9c_user.presence, i9c_user.last_seen' USING in_user_id;
  
  RETURN;
END;
$_$;


ALTER FUNCTION public.edit_user(in_user_id integer, in_col_updates character varying[]) OWNER TO postgres;

--
-- Name: end_signup_session(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.end_signup_session(in_session_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM ongoing_signup 
  WHERE session_id = in_session_id;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.end_signup_session(in_session_id uuid) OWNER TO postgres;

--
-- Name: find_nearby_users(integer, circle); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.find_nearby_users(in_client_id integer, in_user_live_location circle) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT i9c_user.id, i9c_user.username, i9c_user.profile_picture, 
  i9c_user.presence, i9c_user.last_seen 
               FROM i9c_user 
			   WHERE in_user_live_location @> point(i9c_user.location) AND deleted = false AND i9c_user.id != in_client_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.find_nearby_users(in_client_id integer, in_user_live_location circle) OWNER TO postgres;

--
-- Name: get_all_users(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_all_users(in_client_id integer) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
    SELECT i9c_user.id, i9c_user.username, i9c_user.profile_picture, i9c_user.presence, i9c_user.last_seen
	FROM i9c_user 
	WHERE i9c_user.id != in_client_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_all_users(in_client_id integer) OWNER TO postgres;

--
-- Name: get_dm_chat_events_pending_receipt(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_dm_chat_events_pending_receipt(in_user_id integer) RETURNS TABLE(event_data_kvp json)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT json_build_object('event', event, 'data', data) FROM dm_chat_event_pending_receipt
  WHERE user_id = in_user_id ORDER BY created_at;
  
  DELETE FROM dm_chat_event_pending_receipt WHERE user_id = in_user_id;
  RETURN;
END;
$$;


ALTER FUNCTION public.get_dm_chat_events_pending_receipt(in_user_id integer) OWNER TO postgres;

--
-- Name: get_dm_chat_history(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_dm_chat_history(in_dm_chat_id integer) RETURNS TABLE(message json, created_at timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT json_build_object(
	  'id', dcm.id,
	  'sender', json_build_object(
		  'id', sender.id,
		  'username', sender.username,
		  'profile_picture', sender.profile_picture
	  ),
	  'content', dcm.msg_content,
	  'edited', dcm.edited,
	  'delivery_status', dcm.delivery_status,
	  'created_at', dcm.created_at,
	  'edited_at', dcm.edited_at,
	  'reactions', json_agg(
		  json_build_object(
			  'reaction', dcmr.reaction,
			  'reactor', json_build_object(
				  'id', reactor.id,
				  'username', reactor.username,
				  'profile_picture', reactor.profile_picture
			  )
		  )
	  )
  ) AS msg, dcm.created_at FROM dm_chat_message dcm
  INNER JOIN i9c_user sender ON sender.id = dcm.sender_id
  LEFT JOIN dm_chat_message_reaction dcmr ON dcmr.message_id = dcm.id AND dcmr.deleted = false
  INNER JOIN i9c_user reactor ON reactor.id = dcmr.reactor_id
  WHERE dcm.dm_chat_id = in_dm_chat_id AND dcm.deleted = false
  GROUP BY dcm.id, sender.id
  ORDER BY dcm.created_at DESC;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_dm_chat_history(in_dm_chat_id integer) OWNER TO postgres;

--
-- Name: get_dm_chat_message_events_pending_receipt(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_dm_chat_message_events_pending_receipt(in_user_id integer, in_dm_chat_id integer) RETURNS TABLE(event_data_kvp json)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT json_build_object('event', event, 'data', data) FROM dm_chat_message_event_pending_receipt
  WHERE user_id = in_user_id AND dm_chat_id = in_dm_chat_id ORDER BY created_at;
  
  DELETE FROM dm_chat_message_event_pending_receipt WHERE user_id = in_user_id AND dm_chat_id = in_dm_chat_id;
  RETURN;
END;
$$;


ALTER FUNCTION public.get_dm_chat_message_events_pending_receipt(in_user_id integer, in_dm_chat_id integer) OWNER TO postgres;

--
-- Name: get_group_chat_events_pending_receipt(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_group_chat_events_pending_receipt(in_user_id integer) RETURNS TABLE(event_data_kvp json)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT json_build_object('event', event, 'data', data) FROM group_chat_event_pending_receipt
  WHERE user_id = in_user_id ORDER BY created_at;
  
  DELETE FROM group_chat_event_pending_receipt WHERE user_id = in_user_id;
  RETURN;
END;
$$;


ALTER FUNCTION public.get_group_chat_events_pending_receipt(in_user_id integer) OWNER TO postgres;

--
-- Name: get_group_chat_history(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_group_chat_history(in_group_chat_id integer) RETURNS TABLE(history_item jsonb, time_created timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT jsonb_build_object(
	  'type', 'message',
	  'id', gcm.id,
	  'sender', jsonb_build_object(
		  'id', sender.id,
		  'username', sender.username,
		  'profile_picture', sender.profile_picture
	  ),
	  'content', gcm.msg_content,
	  'edited', gcm.edited,
	  'delivery_status', gcm.delivery_status,
	  'created_at', gcm.created_at,
	  'edited_at', gcm.edited_at,
	  'reactions', jsonb_agg(
		  jsonb_build_object(
			  'reaction', gcmr.reaction,
			  'reactor', jsonb_build_object(
				  'id', reactor.id,
				  'username', reactor.username,
				  'profile_picture', reactor.profile_picture
			  )
		  )
	  )
  ) AS hitem, gcm.created_at FROM group_chat_message gcm
  INNER JOIN i9c_user sender ON sender.id = gcm.sender_id
  LEFT JOIN group_chat_message_reaction gcmr ON gcmr.message_id = gcm.id AND gcmr.deleted = false
  INNER JOIN i9c_user reactor ON reactor.id = gcmr.reactor_id
  WHERE gcm.group_chat_id = in_group_chat_id AND gcm.deleted = false
  GROUP BY gcm.id, sender.id
  UNION
  SELECT jsonb_build_object(
	  'type', 'activity',
	  'activity_type', gcal.activity_type,
	  'activity_info', gcal.activity_info
  ) AS hitem, gcal.created_at FROM group_chat_activity_log gcal
  WHERE group_chat_id = in_group_chat_id
  ORDER BY created_at DESC;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_group_chat_history(in_group_chat_id integer) OWNER TO postgres;

--
-- Name: get_group_chat_message_events_pending_receipt(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_group_chat_message_events_pending_receipt(in_user_id integer, in_group_chat_id integer) RETURNS TABLE(event_data_kvp json)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
  SELECT json_build_object('event', event, 'data', data) FROM group_chat_message_event_pending_receipt
  WHERE user_id = in_user_id AND group_chat_id = in_group_chat_id ORDER BY created_at;
  
  DELETE FROM group_chat_message_event_pending_receipt WHERE user_id = in_user_id AND group_chat_id = in_group_chat_id;
  RETURN;
END;
$$;


ALTER FUNCTION public.get_group_chat_message_events_pending_receipt(in_user_id integer, in_group_chat_id integer) OWNER TO postgres;

--
-- Name: get_my_chats(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_my_chats(in_user_id integer) RETURNS TABLE(chat jsonb, last_active timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY
  SELECT jsonb_build_object(
	  'type', 'dm',
	  'id', dm_chat_id,
	  'updated_at', updated_at,
	  'unread_messages_count', unread_messages_count,
	  'partner', jsonb_build_object(
		  'id', partner.id,
		  'username', partner.username,
		  'profile_picture', partner.profile_picture,
		  'presence', partner.presence,
		  'last_seen', partner.last_seen
	  )
  ) AS my_chat, updated_at FROM user_dm_chat
  INNER JOIN i9c_user partner ON partner.id = user_dm_chat.partner_id
  WHERE user_dm_chat.user_id = in_user_id
  UNION
  SELECT jsonb_build_object(
	  'type', 'group',
	  'id', group_chat_id,
	  'name', group_chat.name,
	  'picture', group_chat.picture,
	  'updated_at', updated_at,
	  'unread_messages_count', unread_messages_count
  ) AS my_chat, updated_at FROM user_group_chat
  INNER JOIN group_chat ON group_chat.id = user_group_chat.group_chat_id
  WHERE user_group_chat.user_id = in_user_id
  ORDER BY updated_at DESC;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_my_chats(in_user_id integer) OWNER TO postgres;

--
-- Name: get_user(anycompatible); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user(unique_identifier anycompatible) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY SELECT i9c_user.id, i9c_user.username, i9c_user.profile_picture, i9c_user.presence, i9c_user.last_seen FROM i9c_user 
  WHERE unique_identifier::varchar = ANY(ARRAY[i9c_user.id::varchar, i9c_user.email, i9c_user.username]) AND deleted = false;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.get_user(unique_identifier anycompatible) OWNER TO postgres;

--
-- Name: get_user_password(anycompatible); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_user_password(unique_identifier anycompatible, OUT password character varying) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT i9c_user.password FROM i9c_user 
  WHERE unique_identifier::varchar = ANY(ARRAY[i9c_user.id::varchar, i9c_user.email, i9c_user.username]) AND deleted = false 
  INTO "password";
END;
$$;


ALTER FUNCTION public.get_user_password(unique_identifier anycompatible, OUT password character varying) OWNER TO postgres;

--
-- Name: join_group(integer, character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.join_group(in_group_chat_id integer, in_user character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN 
  -- create user_group_chat for in_user
  INSERT INTO user_group_chat (user_id, group_chat_id)
  VALUES (in_user[1]::int, in_group_chat_id);
    
  -- create group_chat_membership for in_user
  INSERT INTO group_chat_membership (group_chat_id, member_id)
  VALUES (in_group_chat_id, in_user[1]::int);
  
  -- increment members_count in group_chat by 1
  UPDATE group_chat 
  SET members_count = members_count + 1
  WHERE id = in_group_chat_id;
    
  -- create group_chat_activity_log for user joined
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'user_joined', json_build_object('username', in_user[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.join_group(in_group_chat_id integer, in_user character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: leave_group(integer, character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.leave_group(in_group_chat_id integer, in_user character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN 
  -- set user_group_chat "deleted = true" for user
  UPDATE user_group_chat 
  SET deleted = true
  WHERE user_id = in_user[1] AND group_chat_id = in_group_chat_id;
  
  -- set group_chat_membership "deleted = true" for user
  UPDATE group_chat_membership 
  SET deleted = true
  WHERE group_chat_id = in_group_chat_id AND member_id = in_user[1];
  
  -- decrement group_chat members_count by 1
  UPDATE group_chat
  SET members_count = members_count - 1
  WHERE id = in_group_chat_id;
    
  -- create group_chat_activity_log for user left
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'user_left', json_build_object('username', in_user[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.leave_group(in_group_chat_id integer, in_user character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: make_user_group_admin(integer, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.make_user_group_admin(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set group_chat_membership, "role = 'admin'" for user
  UPDATE group_chat_membership
  SET "role" = 'admin'
  WHERE group_chat_id = in_group_chat_id AND member_id = in_user[1];
  
  -- create group_chat_activity_log for make admin
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'user_made_group_admin', json_build_object('made_by', in_admin[2], 'username', in_user[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.make_user_group_admin(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: new_dm_chat(integer, integer, json, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_dm_chat(in_initiator_id integer, in_partner_id integer, in_init_msg_content json, in_created_at timestamp without time zone, OUT initiator_resp_data json, OUT partner_resp_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  new_dm_chat_id int;
  init_msg_id int;
  
  initiator_user_data json;
BEGIN
  -- create dm_chat
  INSERT INTO dm_chat (initiator_id, partner_id, created_at) 
  VALUES (in_initiator_id, in_partner_id, in_created_at)
  RETURNING id INTO new_dm_chat_id;
  
  -- create user_dm_chat for both initiator and partner
  INSERT INTO user_dm_chat (user_id, partner_id, dm_chat_id, updated_at) 
  VALUES (in_initiator_id, in_partner_id, new_dm_chat_id, in_created_at);
  
  INSERT INTO user_dm_chat (user_id, partner_id, dm_chat_id) 
  VALUES (in_partner_id, in_initiator_id, new_dm_chat_id);
  
  -- create dm_chat_message
  INSERT INTO dm_chat_message (sender_id, dm_chat_id, msg_content, created_at) 
  VALUES (in_initiator_id, new_dm_chat_id, in_init_msg_content, in_created_at)
  RETURNING id INTO init_msg_id;
  
  SELECT json_build_object(
		  'id', id,
		  'username', username,
		  'profile_picture', profile_picture,
		  'presence', presence,
		  'last_seen', last_seen
	  ) FROM i9c_user WHERE id = in_initiator_id INTO initiator_user_data;
  
  initiator_resp_data := json_build_object('new_dm_chat_id', new_dm_chat_id, 'init_msg_id', init_msg_id);
  partner_resp_data := json_build_object(
	  'type', 'dm', 
	  'id', new_dm_chat_id, 
	  'partner', initiator_user_data,
	  'init_msg_id', init_msg_id
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_dm_chat(in_initiator_id integer, in_partner_id integer, in_init_msg_content json, in_created_at timestamp without time zone, OUT initiator_resp_data json, OUT partner_resp_data json) OWNER TO postgres;

--
-- Name: new_group_chat(character varying, character varying, character varying, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_group_chat(in_name character varying, in_description character varying, in_picture character varying, in_creator character varying[], in_init_users character varying[], OUT creator_resp_data json, OUT new_members_resp_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  new_group_chat_id int;
  
  iuser varchar[];
  iusers_usernames varchar[];
BEGIN
  -- create group chat
  INSERT INTO group_chat (name, description, picture, creator_id, members_count)
  VALUES (in_name, in_description, in_picture, in_creator[1]::int, array_length(in_init_users, 1))
  RETURNING id INTO new_group_chat_id;
  
  -- input group_chat_activity_log for group created
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (new_group_chat_id, 'group_created', json_build_object('creator', in_creator[2], 'group_name', in_name));
  
  FOREACH iuser SLICE 1 IN ARRAY in_init_users 
  LOOP
    -- create user_group_chat for each iuser
    INSERT INTO user_group_chat (user_id, group_chat_id)
	VALUES (iuser[1]::int, new_group_chat_id);
	
    -- create group_chat_membership for all iusers: set creator's role to admin
	INSERT INTO group_chat_membership (group_chat_id, member_id, "role")
    VALUES (new_group_chat_id, iuser[1]::int, CASE WHEN iuser[1]::int = in_creator[1]::int THEN 'admin' ELSE 'member' END);
	
	-- extract each iuser's username for later
	iusers_usernames := array_append(iusers_usernames, iuser[2]);
  END LOOP;
  
  -- input group_chat_activity_log for users added
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (new_group_chat_id, 'users_added', json_build_object('added_by', in_creator[2], 'new_users', iusers_usernames));
  
  creator_resp_data := json_build_object('new_group_chat_id', new_group_chat_id);
  new_members_resp_data := json_build_object(
	  'type', 'group',
	  'id', new_group_chat_id,
	  'name', in_name,
	  'picture', in_picture,
	  'updated_at', now()::timestamp,
	  'unread_messages_count', 0
  )
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_group_chat(in_name character varying, in_description character varying, in_picture character varying, in_creator character varying[], in_init_users character varying[], OUT creator_resp_data json, OUT new_members_resp_data json) OWNER TO postgres;

--
-- Name: new_signup_session(character varying, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_signup_session(in_email character varying, in_verification_code integer, OUT session_id uuid) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
BEGIN
  DELETE FROM ongoing_signup WHERE email = in_email;  
  
  INSERT INTO ongoing_signup (email, verification_code)
  VALUES (in_email, in_verification_code)
  RETURNING ongoing_signup.session_id INTO session_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_signup_session(in_email character varying, in_verification_code integer, OUT session_id uuid) OWNER TO postgres;

--
-- Name: new_user(character varying, character varying, character varying, circle); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying, in_location circle) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY INSERT INTO i9c_user (email, username, password, location) 
  VALUES (in_email, in_username, in_password, in_location)
  RETURNING i9c_user.id, i9c_user.username, i9c_user.profile_picture, i9c_user.presence, i9c_user.last_seen;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.new_user(in_email character varying, in_username character varying, in_password character varying, in_location circle) OWNER TO postgres;

--
-- Name: react_to_dm_chat_message(integer, integer, integer, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.react_to_dm_chat_message(in_dm_chat_id integer, in_msg_id integer, in_reactor_id integer, in_reaction character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
 INSERT INTO dm_chat_message_reaction (dm_chat_id, message_id, reactor_id, reaction)
 VALUES (in_dm_chat_id, in_msg_id, in_reactor_id, in_reaction);
 
 RETURN true;
END;
$$;


ALTER FUNCTION public.react_to_dm_chat_message(in_dm_chat_id integer, in_msg_id integer, in_reactor_id integer, in_reaction character varying) OWNER TO postgres;

--
-- Name: react_to_group_chat_message(integer, integer, integer, character varying); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.react_to_group_chat_message(in_group_chat_id integer, in_msg_id integer, in_reactor_id integer, in_reaction character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
 INSERT INTO group_chat_message_reaction (group_chat_id, message_id, reactor_id, reaction)
 VALUES (in_group_chat_id, in_msg_id, in_reactor_id, in_reaction);
 
 RETURN true;
END;
$$;


ALTER FUNCTION public.react_to_group_chat_message(in_group_chat_id integer, in_msg_id integer, in_reactor_id integer, in_reaction character varying) OWNER TO postgres;

--
-- Name: remove_user_from_group(integer, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.remove_user_from_group(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set user_group_chat "deleted = true" for user
  UPDATE user_group_chat 
  SET deleted = true
  WHERE user_id = in_user[1] AND group_chat_id = in_group_chat_id;
  
  -- set group_chat_membership "deleted = true" for user
  UPDATE group_chat_membership 
  SET deleted = true
  WHERE group_chat_id = in_group_chat_id AND member_id = in_user[1];
  
  -- decrement group_chat members_count by 1
  UPDATE group_chat
  SET members_count = members_count - 1
  WHERE id = in_group_chat_id;
  
  -- create group_chat_activity_log for the user removed
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'user_removed', json_build_object('removed_by', in_admin[2], 'username', in_user[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.remove_user_from_group(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: remove_user_from_group_admins(integer, character varying[], character varying[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.remove_user_from_group_admins(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) RETURNS record
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- check if in_admin_id is an admin in the group, if not, raise an exception
  IF (SELECT NOT EXISTS(SELECT 1 
					   FROM group_chat_membership 
					   WHERE group_chat_id = in_group_chat_id AND member_id = in_admin[1] AND "role" = 'admin')
  ) THEN
    RAISE EXCEPTION 'user (id=%) is not a group admin', in_admin[1];
  END IF;
  
  -- set group_chat_membership, "role = 'member'" for user
  UPDATE group_chat_membership
  SET "role" = 'member'
  WHERE group_chat_id = in_group_chat_id AND member_id = in_user[1];
  
  -- create group_chat_activity_log for remove admin
  INSERT INTO group_chat_activity_log (group_chat_id, activity_type, activity_info)
  VALUES (in_group_chat_id, 'user_removed_from_group_admins', json_build_object('removed_by', in_admin[2], 'username', in_user[2]))
  RETURNING json_build_object('group_chat_id', group_chat_id, 'activity_type', activity_type, 'activity_info', activity_info) INTO activity_data;
  
  SELECT array_agg(member_id) FROM group_chat_membership 
  WHERE group_chat_id = in_group_chat_id AND member_id != in_admin[1] AND deleted = false
  INTO member_ids;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.remove_user_from_group_admins(in_group_chat_id integer, in_admin character varying[], in_user character varying[], OUT member_ids integer[], OUT activity_data json) OWNER TO postgres;

--
-- Name: search_user(integer, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.search_user(in_client_id integer, search_query text) RETURNS TABLE(id integer, username character varying, profile_picture character varying, presence character varying, last_seen timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
  RETURN QUERY 
    SELECT i9c_user.id, i9c_user.username, i9c_user.profile_picture, i9c_user.presence, i9c_user.last_seen
	FROM i9c_user 
	WHERE i9c_user.username LIKE format('%%%s%%', search_query) AND deleted = false AND i9c_user.id != in_client_id;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.search_user(in_client_id integer, search_query text) OWNER TO postgres;

--
-- Name: send_dm_chat_message(integer, integer, json, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.send_dm_chat_message(in_dm_chat_id integer, in_sender_id integer, in_msg_content json, in_created_at timestamp without time zone, OUT sender_resp_data json, OUT receiver_resp_data json, OUT receiver_id integer) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  new_msg_id int;
  
  sender_info json;
BEGIN
  -- create dm_chat_message
  INSERT INTO dm_chat_message (sender_id, dm_chat_id, msg_content, created_at) 
  VALUES (in_sender_id, in_dm_chat_id, in_msg_content, in_created_at)
  RETURNING id INTO new_msg_id;
  
  -- update user_dm_chat
  UPDATE user_dm_chat 
  SET updated_at = in_created_at
  WHERE user_id = in_sender_id AND dm_chat_id = in_dm_chat_id
  RETURNING partner_id INTO receiver_id;
  
  SELECT json_build_object (
	  'id', id,
	  'username', username,
	  'profile_picture', profile_picture
  ) FROM i9c_user WHERE id = in_sender_id INTO sender_info;
  
  sender_resp_data := json_build_object('new_msg_id', new_msg_id);
  receiver_resp_data := json_build_object(
	  'id', new_msg_id,
	  'dm_chat_id', in_dm_chat_id,
	  'sender', sender_info,
	  'content', in_msg_content,
	  'reactions', ARRAY[]::json[]
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.send_dm_chat_message(in_dm_chat_id integer, in_sender_id integer, in_msg_content json, in_created_at timestamp without time zone, OUT sender_resp_data json, OUT receiver_resp_data json, OUT receiver_id integer) OWNER TO postgres;

--
-- Name: send_group_chat_message(integer, integer, json, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.send_group_chat_message(in_group_chat_id integer, in_sender_id integer, in_msg_content json, in_created_at timestamp without time zone, OUT sender_resp_data json, OUT members_resp_data json, OUT member_ids integer[]) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  new_msg_id int;
  
  sender_info json;
BEGIN
  -- create group_chat_message
  INSERT INTO group_chat_message (sender_id, group_chat_id, msg_content, created_at) 
  VALUES (in_sender_id, in_group_chat_id, in_msg_content, in_created_at)
  RETURNING id INTO new_msg_id;
  
  -- update user_group_chat
  UPDATE user_group_chat 
  SET updated_at = in_created_at
  WHERE user_id = in_sender_id AND group_chat_id = in_group_chat_id;
  
  -- log message delivery as seen
  INSERT INTO group_chat_message_delivery (group_chat_id, message_id, user_id, status)
  VALUES (in_group_chat_id, new_msg_id, in_sender_id, 'seen');
  
  SELECT array_agg(member_id) FROM group_chat_membership
  WHERE member_id != in_sender_id AND group_chat_id = in_group_chat_id AND deleted = false
  INTO member_ids;
  
  SELECT json_build_object (
	  'id', id,
	  'username', username,
	  'profile_picture', profile_picture
  ) FROM i9c_user WHERE id = in_sender_id INTO sender_info;
  
  sender_resp_data := json_build_object('new_msg_id', new_msg_id);
  members_resp_data := json_build_object(
	  'id', new_msg_id,
	  'group_chat_id', in_group_chat_id,
	  'sender', sender_info,
	  'content', in_msg_content,
	  'reactions', ARRAY[]::json[]
  );
  
  RETURN;
END;
$$;


ALTER FUNCTION public.send_group_chat_message(in_group_chat_id integer, in_sender_id integer, in_msg_content json, in_created_at timestamp without time zone, OUT sender_resp_data json, OUT members_resp_data json, OUT member_ids integer[]) OWNER TO postgres;

--
-- Name: signup_session_email_verified(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.signup_session_email_verified(in_session_id uuid, OUT is_verified boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  SELECT verified FROM ongoing_signup 
  WHERE session_id = in_session_id
  INTO is_verified;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.signup_session_email_verified(in_session_id uuid, OUT is_verified boolean) OWNER TO postgres;

--
-- Name: switch_user_presence(integer, character varying, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.switch_user_presence(in_user_id integer, in_presence character varying, in_last_seen timestamp without time zone) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
declare 
ls timestamp := null;
begin
if in_presence not in ('offline', 'online') then
raise exception 'Invalid presence value "%"', in_presence;
end if;

if in_presence = 'offline' then
ls := in_last_seen;
end if;

update i9c_user set presence = in_presence, last_seen = ls where id = in_user_id;

return true;
end;
$$;


ALTER FUNCTION public.switch_user_presence(in_user_id integer, in_presence character varying, in_last_seen timestamp without time zone) OWNER TO postgres;

--
-- Name: update_dm_chat_message_delivery_status(integer, integer, integer, character varying, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_dm_chat_message_delivery_status(in_dm_chat_id integer, in_msg_id integer, in_receiver_id integer, in_status character varying, in_updated_at timestamp without time zone) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
 UPDATE dm_chat_message 
 SET delivery_status = in_status
 WHERE id = in_msg_id AND dm_chat_id = in_dm_chat_id;
 
 IF in_status = 'delivered' THEN
   UPDATE user_dm_chat 
   SET updated_at = in_updated_at, unread_messages_count = unread_messages_count + 1 
   WHERE dm_chat_id = in_dm_chat_id AND user_id = in_receiver_id;
 ELSIF in_status = 'seen' THEN
   UPDATE user_dm_chat 
   SET unread_messages_count = CASE WHEN (unread_messages_count - 1) < 0 THEN 0 ELSE unread_messages_count - 1 END
   WHERE dm_chat_id = in_dm_chat_id AND user_id = in_receiver_id;
 ELSE
   RAISE EXCEPTION 'Invalid update value, "%"', in_status;
 END IF;
 
 RETURN true;
END;
$$;


ALTER FUNCTION public.update_dm_chat_message_delivery_status(in_dm_chat_id integer, in_msg_id integer, in_receiver_id integer, in_status character varying, in_updated_at timestamp without time zone) OWNER TO postgres;

--
-- Name: update_group_chat_message_delivery_status(integer, integer, integer, character varying, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_group_chat_message_delivery_status(in_group_chat_id integer, in_msg_id integer, in_receiver_id integer, in_status character varying, in_updated_at timestamp without time zone, OUT overall_delivery_status character varying, OUT should_broadcast boolean) RETURNS record
    LANGUAGE plpgsql
    AS $$
DECLARE
  group_chat_members_count int;
  delivered_to_count int;
  seen_by_count int;
  
  old_delivery_status varchar;
  new_delivery_status varchar;
BEGIN
  SELECT members_count FROM group_chat 
  WHERE id = in_group_chat_id
  INTO group_chat_members_count;
   
  SELECT delivery_status, delivery_status FROM group_chat_message 
  WHERE id = in_msg_id AND group_chat_id = in_group_chat_id
  INTO old_delivery_status, new_delivery_status;

 IF in_status = 'delivered' THEN
   INSERT INTO group_chat_message_delivery (group_chat_id, message_id, user_id)
   VALUES (in_group_chat_id, in_msg_id, in_receiver_id);
 
   UPDATE user_group_chat 
   SET updated_at = in_updated_at, unread_messages_count = unread_messages_count + 1 
   WHERE group_chat_id = in_group_chat_id AND user_id = in_receiver_id;
   
   SELECT COUNT(1) FROM group_chat_message_delivery 
   WHERE message_id = in_msg_id AND group_chat_id = in_group_chat_id AND status = 'delivered'
   INTO delivered_to_count;
   
   IF group_chat_members_count = delivered_to_count THEN
     UPDATE group_chat_message 
	 SET delivery_status = 'delivered'
	 WHERE id = in_msg_id AND group_chat_id = in_group_chat_id;
	 
	 new_delivery_status := 'delivered';
   END IF;
 ELSIF in_status = 'seen' THEN
   UPDATE group_chat_message_delivery 
   SET status = 'seen'
   WHERE message_id = in_msg_id AND user_id = in_receiver_id AND group_chat_id = in_group_chat_id;
 
   UPDATE user_group_chat 
   SET unread_messages_count = CASE WHEN (unread_messages_count - 1) < 0 THEN 0 ELSE unread_messages_count - 1 END
   WHERE group_chat_id = in_group_chat_id AND user_id = in_receiver_id;
   
   SELECT COUNT(1) FROM group_chat_message_delivery 
   WHERE message_id = in_msg_id AND group_chat_id = in_group_chat_id AND status = 'seen'
   INTO seen_by_count;
   
   IF group_chat_members_count = seen_by_count THEN
     UPDATE group_chat_message 
	 SET delivery_status = 'seen'
	 WHERE id = in_msg_id AND group_chat_id = in_group_chat_id;
	 
	 new_delivery_status := 'seen';
   END IF;
 ELSE
   RAISE EXCEPTION 'Invalid update value, "%"', in_status;
 END IF;
 
 overall_delivery_status := new_delivery_status;
 should_broadcast := old_delivery_status != new_delivery_status;
 
 RETURN;
END;
$$;


ALTER FUNCTION public.update_group_chat_message_delivery_status(in_group_chat_id integer, in_msg_id integer, in_receiver_id integer, in_status character varying, in_updated_at timestamp without time zone, OUT overall_delivery_status character varying, OUT should_broadcast boolean) OWNER TO postgres;

--
-- Name: update_user_location(integer, circle); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_user_location(in_user_id integer, in_new_location circle) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE i9c_user 
  SET "location" = in_new_location
  WHERE id = in_user_id;
  
  RETURN true;
END;
$$;


ALTER FUNCTION public.update_user_location(in_user_id integer, in_new_location circle) OWNER TO postgres;

--
-- Name: verify_email(uuid, integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.verify_email(in_session_id uuid, in_verf_code integer, OUT is_success boolean) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN    
  IF (SELECT verification_code FROM ongoing_signup WHERE session_id = in_session_id) = in_verf_code THEN
    UPDATE ongoing_signup SET verified = true 
	WHERE session_id = in_session_id;
    is_success := true;
  ELSE 
    is_success := false;
  END IF;
  
  RETURN;
END;
$$;


ALTER FUNCTION public.verify_email(in_session_id uuid, in_verf_code integer, OUT is_success boolean) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: dm_chat; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_chat (
    id integer NOT NULL,
    initiator_id integer NOT NULL,
    partner_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE public.dm_chat OWNER TO postgres;

--
-- Name: dm_chat_event_pending_receipt; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_chat_event_pending_receipt (
    id integer NOT NULL,
    user_id integer NOT NULL,
    event character varying NOT NULL,
    data json NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.dm_chat_event_pending_receipt OWNER TO postgres;

--
-- Name: dm_chat_event_pending_receipt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dm_chat_event_pending_receipt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dm_chat_event_pending_receipt_id_seq OWNER TO postgres;

--
-- Name: dm_chat_event_pending_receipt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dm_chat_event_pending_receipt_id_seq OWNED BY public.dm_chat_event_pending_receipt.id;


--
-- Name: dm_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dm_chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dm_chat_id_seq OWNER TO postgres;

--
-- Name: dm_chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dm_chat_id_seq OWNED BY public.dm_chat.id;


--
-- Name: dm_chat_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_chat_message (
    id integer NOT NULL,
    sender_id integer NOT NULL,
    dm_chat_id integer NOT NULL,
    msg_content json NOT NULL,
    edited boolean DEFAULT false NOT NULL,
    delivery_status character varying DEFAULT 'sent'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    edited_at timestamp without time zone,
    deleted boolean DEFAULT false NOT NULL,
    CONSTRAINT dm_chat_message_delivery_status_check CHECK (((delivery_status)::text = ANY (ARRAY['sent'::text, 'delivered'::text, 'seen'::text])))
);


ALTER TABLE public.dm_chat_message OWNER TO postgres;

--
-- Name: dm_chat_message_event_pending_receipt; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_chat_message_event_pending_receipt (
    id integer NOT NULL,
    user_id integer NOT NULL,
    dm_chat_id integer NOT NULL,
    event character varying NOT NULL,
    data json NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.dm_chat_message_event_pending_receipt OWNER TO postgres;

--
-- Name: dm_chat_message_event_pending_receipt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dm_chat_message_event_pending_receipt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dm_chat_message_event_pending_receipt_id_seq OWNER TO postgres;

--
-- Name: dm_chat_message_event_pending_receipt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dm_chat_message_event_pending_receipt_id_seq OWNED BY public.dm_chat_message_event_pending_receipt.id;


--
-- Name: dm_chat_message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dm_chat_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dm_chat_message_id_seq OWNER TO postgres;

--
-- Name: dm_chat_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dm_chat_message_id_seq OWNED BY public.dm_chat_message.id;


--
-- Name: dm_chat_message_reaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dm_chat_message_reaction (
    id integer NOT NULL,
    message_id integer NOT NULL,
    reactor_id integer NOT NULL,
    reaction character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    dm_chat_id integer NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE public.dm_chat_message_reaction OWNER TO postgres;

--
-- Name: dm_chat_message_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.dm_chat_message_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dm_chat_message_reaction_id_seq OWNER TO postgres;

--
-- Name: dm_chat_message_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.dm_chat_message_reaction_id_seq OWNED BY public.dm_chat_message_reaction.id;


--
-- Name: group_chat; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat (
    id integer NOT NULL,
    creator_id integer NOT NULL,
    name character varying NOT NULL,
    description character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    members_count integer NOT NULL,
    picture character varying DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.group_chat OWNER TO postgres;

--
-- Name: group_chat_activity_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_activity_log (
    id integer NOT NULL,
    group_chat_id integer NOT NULL,
    activity_type character varying NOT NULL,
    activity_info json NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.group_chat_activity_log OWNER TO postgres;

--
-- Name: group_chat_activity_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_activity_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_activity_log_id_seq OWNER TO postgres;

--
-- Name: group_chat_activity_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_activity_log_id_seq OWNED BY public.group_chat_activity_log.id;


--
-- Name: group_chat_event_pending_receipt; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_event_pending_receipt (
    id integer NOT NULL,
    user_id integer NOT NULL,
    event character varying NOT NULL,
    data json NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.group_chat_event_pending_receipt OWNER TO postgres;

--
-- Name: group_chat_event_pending_receipt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_event_pending_receipt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_event_pending_receipt_id_seq OWNER TO postgres;

--
-- Name: group_chat_event_pending_receipt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_event_pending_receipt_id_seq OWNED BY public.group_chat_event_pending_receipt.id;


--
-- Name: group_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_id_seq OWNER TO postgres;

--
-- Name: group_chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_id_seq OWNED BY public.group_chat.id;


--
-- Name: group_chat_membership; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_membership (
    id integer NOT NULL,
    group_chat_id integer NOT NULL,
    member_id integer NOT NULL,
    role character varying DEFAULT 'member'::character varying NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    CONSTRAINT group_chat_membership_role_check CHECK (((role)::text = ANY (ARRAY['admin'::text, 'member'::text])))
);


ALTER TABLE public.group_chat_membership OWNER TO postgres;

--
-- Name: group_chat_membership_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_membership_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_membership_id_seq OWNER TO postgres;

--
-- Name: group_chat_membership_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_membership_id_seq OWNED BY public.group_chat_membership.id;


--
-- Name: group_chat_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_message (
    id integer NOT NULL,
    sender_id integer NOT NULL,
    group_chat_id integer NOT NULL,
    msg_content json NOT NULL,
    edited boolean DEFAULT false NOT NULL,
    delivery_status character varying DEFAULT 'sent'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    edited_at timestamp without time zone,
    deleted boolean DEFAULT false NOT NULL,
    CONSTRAINT group_chat_message_delivery_status_check CHECK (((delivery_status)::text = ANY (ARRAY['sent'::text, 'delivered'::text, 'seen'::text])))
);


ALTER TABLE public.group_chat_message OWNER TO postgres;

--
-- Name: group_chat_message_delivery; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_message_delivery (
    id integer NOT NULL,
    group_chat_id integer NOT NULL,
    message_id integer NOT NULL,
    user_id integer NOT NULL,
    status character varying DEFAULT 'delivered'::character varying NOT NULL,
    CONSTRAINT group_chat_message_delivery_status_check1 CHECK (((status)::text = ANY (ARRAY['delivered'::text, 'seen'::text])))
);


ALTER TABLE public.group_chat_message_delivery OWNER TO postgres;

--
-- Name: group_chat_message_delivery_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_message_delivery_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_message_delivery_id_seq OWNER TO postgres;

--
-- Name: group_chat_message_delivery_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_message_delivery_id_seq OWNED BY public.group_chat_message_delivery.id;


--
-- Name: group_chat_message_event_pending_receipt; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_message_event_pending_receipt (
    id integer NOT NULL,
    user_id integer NOT NULL,
    group_chat_id integer NOT NULL,
    event character varying NOT NULL,
    data json NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.group_chat_message_event_pending_receipt OWNER TO postgres;

--
-- Name: group_chat_message_event_pending_receipt_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_message_event_pending_receipt_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_message_event_pending_receipt_id_seq OWNER TO postgres;

--
-- Name: group_chat_message_event_pending_receipt_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_message_event_pending_receipt_id_seq OWNED BY public.group_chat_message_event_pending_receipt.id;


--
-- Name: group_chat_message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_message_id_seq OWNER TO postgres;

--
-- Name: group_chat_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_message_id_seq OWNED BY public.group_chat_message.id;


--
-- Name: group_chat_message_reaction; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.group_chat_message_reaction (
    id integer NOT NULL,
    message_id integer NOT NULL,
    reactor_id integer NOT NULL,
    reaction character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    group_chat_id integer NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE public.group_chat_message_reaction OWNER TO postgres;

--
-- Name: group_chat_message_reaction_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.group_chat_message_reaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.group_chat_message_reaction_id_seq OWNER TO postgres;

--
-- Name: group_chat_message_reaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.group_chat_message_reaction_id_seq OWNED BY public.group_chat_message_reaction.id;


--
-- Name: i9c_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.i9c_user (
    id integer NOT NULL,
    username character varying NOT NULL,
    password character varying NOT NULL,
    email character varying NOT NULL,
    profile_picture character varying DEFAULT ''::character varying NOT NULL,
    presence character varying DEFAULT 'online'::character varying NOT NULL,
    last_seen timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    location circle,
    CONSTRAINT i9c_user_presence_check CHECK (((presence)::text = ANY (ARRAY['online'::text, 'offline'::text])))
);


ALTER TABLE public.i9c_user OWNER TO postgres;

--
-- Name: i9c_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.i9c_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.i9c_user_id_seq OWNER TO postgres;

--
-- Name: i9c_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.i9c_user_id_seq OWNED BY public.i9c_user.id;


--
-- Name: ongoing_signup; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ongoing_signup (
    session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying NOT NULL,
    verification_code integer NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public.ongoing_signup OWNER TO postgres;

--
-- Name: user_dm_chat; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_dm_chat (
    id integer NOT NULL,
    user_id integer NOT NULL,
    partner_id integer NOT NULL,
    dm_chat_id integer NOT NULL,
    unread_messages_count integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE public.user_dm_chat OWNER TO postgres;

--
-- Name: user_dm_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_dm_chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_dm_chat_id_seq OWNER TO postgres;

--
-- Name: user_dm_chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_dm_chat_id_seq OWNED BY public.user_dm_chat.id;


--
-- Name: user_group_chat; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_group_chat (
    id integer NOT NULL,
    user_id integer NOT NULL,
    group_chat_id integer NOT NULL,
    unread_messages_count integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


ALTER TABLE public.user_group_chat OWNER TO postgres;

--
-- Name: user_group_chat_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_group_chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_group_chat_id_seq OWNER TO postgres;

--
-- Name: user_group_chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_group_chat_id_seq OWNED BY public.user_group_chat.id;


--
-- Name: dm_chat id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat ALTER COLUMN id SET DEFAULT nextval('public.dm_chat_id_seq'::regclass);


--
-- Name: dm_chat_event_pending_receipt id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_event_pending_receipt ALTER COLUMN id SET DEFAULT nextval('public.dm_chat_event_pending_receipt_id_seq'::regclass);


--
-- Name: dm_chat_message id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message ALTER COLUMN id SET DEFAULT nextval('public.dm_chat_message_id_seq'::regclass);


--
-- Name: dm_chat_message_event_pending_receipt id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_event_pending_receipt ALTER COLUMN id SET DEFAULT nextval('public.dm_chat_message_event_pending_receipt_id_seq'::regclass);


--
-- Name: dm_chat_message_reaction id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_reaction ALTER COLUMN id SET DEFAULT nextval('public.dm_chat_message_reaction_id_seq'::regclass);


--
-- Name: group_chat id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat ALTER COLUMN id SET DEFAULT nextval('public.group_chat_id_seq'::regclass);


--
-- Name: group_chat_activity_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_activity_log ALTER COLUMN id SET DEFAULT nextval('public.group_chat_activity_log_id_seq'::regclass);


--
-- Name: group_chat_event_pending_receipt id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_event_pending_receipt ALTER COLUMN id SET DEFAULT nextval('public.group_chat_event_pending_receipt_id_seq'::regclass);


--
-- Name: group_chat_membership id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_membership ALTER COLUMN id SET DEFAULT nextval('public.group_chat_membership_id_seq'::regclass);


--
-- Name: group_chat_message id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message ALTER COLUMN id SET DEFAULT nextval('public.group_chat_message_id_seq'::regclass);


--
-- Name: group_chat_message_delivery id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_delivery ALTER COLUMN id SET DEFAULT nextval('public.group_chat_message_delivery_id_seq'::regclass);


--
-- Name: group_chat_message_event_pending_receipt id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_event_pending_receipt ALTER COLUMN id SET DEFAULT nextval('public.group_chat_message_event_pending_receipt_id_seq'::regclass);


--
-- Name: group_chat_message_reaction id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_reaction ALTER COLUMN id SET DEFAULT nextval('public.group_chat_message_reaction_id_seq'::regclass);


--
-- Name: i9c_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9c_user ALTER COLUMN id SET DEFAULT nextval('public.i9c_user_id_seq'::regclass);


--
-- Name: user_dm_chat id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat ALTER COLUMN id SET DEFAULT nextval('public.user_dm_chat_id_seq'::regclass);


--
-- Name: user_group_chat id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group_chat ALTER COLUMN id SET DEFAULT nextval('public.user_group_chat_id_seq'::regclass);


--
-- Name: dm_chat dm_chat_initiator_id_partner_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat
    ADD CONSTRAINT dm_chat_initiator_id_partner_id_key UNIQUE (initiator_id, partner_id);


--
-- Name: dm_chat_message dm_chat_message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message
    ADD CONSTRAINT dm_chat_message_pkey PRIMARY KEY (id);


--
-- Name: dm_chat_message_reaction dm_chat_message_reaction_message_id_reactor_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_reaction
    ADD CONSTRAINT dm_chat_message_reaction_message_id_reactor_id_key UNIQUE (message_id, reactor_id);


--
-- Name: dm_chat dm_chat_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat
    ADD CONSTRAINT dm_chat_pkey PRIMARY KEY (id);


--
-- Name: group_chat_membership group_chat_membership_group_chat_id_member_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_membership
    ADD CONSTRAINT group_chat_membership_group_chat_id_member_id_key UNIQUE (group_chat_id, member_id);


--
-- Name: group_chat_message_delivery group_chat_message_delivery_message_id_user_id_group_chat_i_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_delivery
    ADD CONSTRAINT group_chat_message_delivery_message_id_user_id_group_chat_i_key UNIQUE (message_id, user_id, group_chat_id);


--
-- Name: group_chat_message group_chat_message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message
    ADD CONSTRAINT group_chat_message_pkey PRIMARY KEY (id);


--
-- Name: group_chat_message_reaction group_chat_message_reaction_message_id_reactor_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_reaction
    ADD CONSTRAINT group_chat_message_reaction_message_id_reactor_id_key UNIQUE (message_id, reactor_id);


--
-- Name: group_chat group_chat_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat
    ADD CONSTRAINT group_chat_pkey PRIMARY KEY (id);


--
-- Name: i9c_user i9c_user_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9c_user
    ADD CONSTRAINT i9c_user_email_key UNIQUE (email);


--
-- Name: i9c_user i9c_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9c_user
    ADD CONSTRAINT i9c_user_pkey PRIMARY KEY (id);


--
-- Name: i9c_user i9c_user_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.i9c_user
    ADD CONSTRAINT i9c_user_username_key UNIQUE (username);


--
-- Name: ongoing_signup ongoing_signup_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ongoing_signup
    ADD CONSTRAINT ongoing_signup_pkey PRIMARY KEY (session_id);


--
-- Name: user_dm_chat user_dm_chat_user_id_dm_chat_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat
    ADD CONSTRAINT user_dm_chat_user_id_dm_chat_id_key UNIQUE (user_id, dm_chat_id);


--
-- Name: user_dm_chat user_dm_chat_user_id_partner_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat
    ADD CONSTRAINT user_dm_chat_user_id_partner_id_key UNIQUE (user_id, partner_id);


--
-- Name: user_group_chat user_group_chat_user_id_group_chat_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group_chat
    ADD CONSTRAINT user_group_chat_user_id_group_chat_id_key UNIQUE (user_id, group_chat_id);


--
-- Name: dm_chat_event_pending_receipt dm_chat_event_pending_receipt_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_event_pending_receipt
    ADD CONSTRAINT dm_chat_event_pending_receipt_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: dm_chat dm_chat_initiator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat
    ADD CONSTRAINT dm_chat_initiator_id_fkey FOREIGN KEY (initiator_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message dm_chat_message_dm_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message
    ADD CONSTRAINT dm_chat_message_dm_chat_id_fkey FOREIGN KEY (dm_chat_id) REFERENCES public.dm_chat(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message_event_pending_receipt dm_chat_message_event_pending_receipt_dm_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_event_pending_receipt
    ADD CONSTRAINT dm_chat_message_event_pending_receipt_dm_chat_id_fkey FOREIGN KEY (dm_chat_id) REFERENCES public.dm_chat(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message_event_pending_receipt dm_chat_message_event_pending_receipt_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_event_pending_receipt
    ADD CONSTRAINT dm_chat_message_event_pending_receipt_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message_reaction dm_chat_message_reaction_dm_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_reaction
    ADD CONSTRAINT dm_chat_message_reaction_dm_chat_id_fkey FOREIGN KEY (dm_chat_id) REFERENCES public.dm_chat(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message_reaction dm_chat_message_reaction_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_reaction
    ADD CONSTRAINT dm_chat_message_reaction_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.dm_chat_message(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message_reaction dm_chat_message_reaction_reactor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message_reaction
    ADD CONSTRAINT dm_chat_message_reaction_reactor_id_fkey FOREIGN KEY (reactor_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: dm_chat_message dm_chat_message_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat_message
    ADD CONSTRAINT dm_chat_message_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: dm_chat dm_chat_partner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dm_chat
    ADD CONSTRAINT dm_chat_partner_id_fkey FOREIGN KEY (partner_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_activity_log group_chat_activity_log_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_activity_log
    ADD CONSTRAINT group_chat_activity_log_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat group_chat_creator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat
    ADD CONSTRAINT group_chat_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_event_pending_receipt group_chat_event_pending_receipt_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_event_pending_receipt
    ADD CONSTRAINT group_chat_event_pending_receipt_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_membership group_chat_membership_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_membership
    ADD CONSTRAINT group_chat_membership_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat_membership group_chat_membership_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_membership
    ADD CONSTRAINT group_chat_membership_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_delivery group_chat_message_delivery_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_delivery
    ADD CONSTRAINT group_chat_message_delivery_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_delivery group_chat_message_delivery_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_delivery
    ADD CONSTRAINT group_chat_message_delivery_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.group_chat_message(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_delivery group_chat_message_delivery_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_delivery
    ADD CONSTRAINT group_chat_message_delivery_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_event_pending_receipt group_chat_message_event_pending_receipt_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_event_pending_receipt
    ADD CONSTRAINT group_chat_message_event_pending_receipt_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_event_pending_receipt group_chat_message_event_pending_receipt_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_event_pending_receipt
    ADD CONSTRAINT group_chat_message_event_pending_receipt_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_message group_chat_message_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message
    ADD CONSTRAINT group_chat_message_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_reaction group_chat_message_reaction_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_reaction
    ADD CONSTRAINT group_chat_message_reaction_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_reaction group_chat_message_reaction_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_reaction
    ADD CONSTRAINT group_chat_message_reaction_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.group_chat_message(id) ON DELETE CASCADE;


--
-- Name: group_chat_message_reaction group_chat_message_reaction_reactor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message_reaction
    ADD CONSTRAINT group_chat_message_reaction_reactor_id_fkey FOREIGN KEY (reactor_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: group_chat_message group_chat_message_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.group_chat_message
    ADD CONSTRAINT group_chat_message_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: user_dm_chat user_dm_chat_dm_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat
    ADD CONSTRAINT user_dm_chat_dm_chat_id_fkey FOREIGN KEY (dm_chat_id) REFERENCES public.dm_chat(id) ON DELETE CASCADE;


--
-- Name: user_dm_chat user_dm_chat_partner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat
    ADD CONSTRAINT user_dm_chat_partner_id_fkey FOREIGN KEY (partner_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: user_dm_chat user_dm_chat_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_dm_chat
    ADD CONSTRAINT user_dm_chat_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- Name: user_group_chat user_group_chat_group_chat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group_chat
    ADD CONSTRAINT user_group_chat_group_chat_id_fkey FOREIGN KEY (group_chat_id) REFERENCES public.group_chat(id) ON DELETE CASCADE;


--
-- Name: user_group_chat user_group_chat_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_group_chat
    ADD CONSTRAINT user_group_chat_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.i9c_user(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

